---
name: database-operations
version: 1.0.0
description: 用于设计数据库架构、编写迁移脚本、优化 SQL 查询、解决 N+1 问题、创建索引、配置 PostgreSQL、设置 EF Core、实现缓存、分区表或任何数据库性能问题。
triggers:
  - database
  - schema
  - migration
  - SQL
  - query optimization
  - index
  - PostgreSQL
  - Postgres
  - N+1
  - slow query
  - EXPLAIN
  - partitioning
  - caching
  - Redis
  - connection pool
  - EF Core migration
  - database design
role: specialist
scope: implementation
output-format: code
---

# 数据库操作

全面的数据库设计、迁移和优化专家。改编自 buildwithclaude by Dave Poon (MIT)。

## 角色定义

你是数据库优化专家，专注于 PostgreSQL、查询性能、架构设计和 EF Core 迁移。你坚持先测量后优化，并始终规划回滚程序。

## 核心原则

1. **先测量** — 优化前始终使用 `EXPLAIN ANALYZE`
2. **战略性索引** — 基于查询模式，而非每列都索引
3. **有选择地反规范化** — 仅在读取模式合理时
4. **缓存昂贵计算** — Redis/物化视图用于热点路径
5. **规划回滚** — 每个迁移都有反向迁移
6. **零停机迁移** — 先添加变更，后执行破坏性变更

---

## 架构设计模式

### 用户管理

```sql
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'pending');

CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  username VARCHAR(50) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  status user_status DEFAULT 'active',
  email_verified BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ,  -- Soft delete

  CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
  CONSTRAINT users_names_not_empty CHECK (LENGTH(TRIM(first_name)) > 0 AND LENGTH(TRIM(last_name)) > 0)
);

-- Strategic indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status) WHERE status != 'active';
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
```

### 审计追踪

```sql
CREATE TYPE audit_operation AS ENUM ('INSERT', 'UPDATE', 'DELETE');

CREATE TABLE audit_log (
  id BIGSERIAL PRIMARY KEY,
  table_name VARCHAR(255) NOT NULL,
  record_id BIGINT NOT NULL,
  operation audit_operation NOT NULL,
  old_values JSONB,
  new_values JSONB,
  changed_fields TEXT[],
  user_id BIGINT REFERENCES users(id),
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_table_record ON audit_log(table_name, record_id);
CREATE INDEX idx_audit_user_time ON audit_log(user_id, created_at);

-- Trigger function
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'DELETE' THEN
    INSERT INTO audit_log (table_name, record_id, operation, old_values)
    VALUES (TG_TABLE_NAME, OLD.id, 'DELETE', to_jsonb(OLD));
    RETURN OLD;
  ELSIF TG_OP = 'UPDATE' THEN
    INSERT INTO audit_log (table_name, record_id, operation, old_values, new_values)
    VALUES (TG_TABLE_NAME, NEW.id, 'UPDATE', to_jsonb(OLD), to_jsonb(NEW));
    RETURN NEW;
  ELSIF TG_OP = 'INSERT' THEN
    INSERT INTO audit_log (table_name, record_id, operation, new_values)
    VALUES (TG_TABLE_NAME, NEW.id, 'INSERT', to_jsonb(NEW));
    RETURN NEW;
  END IF;
END;
$$ LANGUAGE plpgsql;

-- Apply to any table
CREATE TRIGGER audit_users
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();
```

### 软删除模式

```sql
-- Query filter view
CREATE VIEW active_users AS SELECT * FROM users WHERE deleted_at IS NULL;

-- Soft delete function
CREATE OR REPLACE FUNCTION soft_delete(p_table TEXT, p_id BIGINT)
RETURNS VOID AS $$
BEGIN
  EXECUTE format('UPDATE %I SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL', p_table)
  USING p_id;
END;
$$ LANGUAGE plpgsql;
```

### 全文搜索

```sql
ALTER TABLE products ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (
    to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(description, '') || ' ' || COALESCE(sku, ''))
  ) STORED;

CREATE INDEX idx_products_search ON products USING gin(search_vector);

-- Query
SELECT * FROM products
WHERE search_vector @@ to_tsquery('english', 'laptop & gaming');
```

---

## 查询优化

### 优化前先分析

```sql
-- Always start here
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT u.id, u.name, COUNT(o.id) as order_count
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > '2024-01-01'
GROUP BY u.id, u.name
ORDER BY order_count DESC;
```

### 索引策略

```sql
-- Single column for exact lookups
CREATE INDEX CONCURRENTLY idx_users_email ON users(email);

-- Composite for multi-column queries (order matters!)
CREATE INDEX CONCURRENTLY idx_orders_user_status ON orders(user_id, status, created_at);

-- Partial index for filtered queries
CREATE INDEX CONCURRENTLY idx_products_low_stock
ON products(inventory_quantity)
WHERE inventory_tracking = true AND inventory_quantity <= 5;

-- Covering index (includes extra columns to avoid table lookup)
CREATE INDEX CONCURRENTLY idx_orders_covering
ON orders(user_id, status) INCLUDE (total, created_at);

-- GIN index for JSONB
CREATE INDEX CONCURRENTLY idx_products_attrs ON products USING gin(attributes);

-- Expression index
CREATE INDEX CONCURRENTLY idx_users_email_lower ON users(lower(email));
```

### 查找未使用的索引

```sql
SELECT
  schemaname, tablename, indexname,
  idx_scan as scans,
  pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;
```

### 查找缺失索引（慢查询）

```sql
-- Enable pg_stat_statements first
SELECT query, calls, total_exec_time, mean_exec_time, rows
FROM pg_stat_statements
WHERE mean_exec_time > 100  -- ms
ORDER BY total_exec_time DESC
LIMIT 20;
```

### N+1 查询检测

```sql
-- Look for repeated similar queries in pg_stat_statements
SELECT query, calls, mean_exec_time
FROM pg_stat_statements
WHERE calls > 100 AND query LIKE '%WHERE%id = $1%'
ORDER BY calls DESC;
```

---

## 迁移模式

### 安全添加列

```sql
-- +migrate Up
-- Always use CONCURRENTLY for indexes in production
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
CREATE INDEX CONCURRENTLY idx_users_phone ON users(phone) WHERE phone IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
```

### 安全重命名列（零停机）

```sql
-- Step 1: Add new column
ALTER TABLE users ADD COLUMN display_name VARCHAR(100);
UPDATE users SET display_name = name;
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;

-- Step 2: Deploy code that writes to both columns
-- Step 3: Deploy code that reads from new column
-- Step 4: Drop old column
ALTER TABLE users DROP COLUMN name;
```

### 表分区

```sql
-- Create partitioned table
CREATE TABLE orders (
  id BIGSERIAL,
  user_id BIGINT NOT NULL,
  total DECIMAL(10,2),
  created_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE orders_2024_01 PARTITION OF orders
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE orders_2024_02 PARTITION OF orders
  FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Auto-create partitions
CREATE OR REPLACE FUNCTION create_monthly_partition(p_table TEXT, p_date DATE)
RETURNS VOID AS $$
DECLARE
  partition_name TEXT := p_table || '_' || to_char(p_date, 'YYYY_MM');
  next_date DATE := p_date + INTERVAL '1 month';
BEGIN
  EXECUTE format(
    'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
    partition_name, p_table, p_date, next_date
  );
END;
$$ LANGUAGE plpgsql;
```

---

## EF Core 迁移 (.NET)

### 创建和应用

```bash
# Add migration
dotnet ef migrations add AddPhoneToUsers -p src/Infrastructure -s src/Api

# Apply
dotnet ef database update -p src/Infrastructure -s src/Api

# Generate idempotent SQL script for production
dotnet ef migrations script -p src/Infrastructure -s src/Api -o migration.sql --idempotent

# Rollback
dotnet ef database update PreviousMigrationName -p src/Infrastructure -s src/Api
```

### EF Core 配置最佳实践

```csharp
// Use AsNoTracking for read queries
var users = await _db.Users
    .AsNoTracking()
    .Where(u => u.Status == UserStatus.Active)
    .Select(u => new UserDto { Id = u.Id, Name = u.Name })
    .ToListAsync(ct);

// Avoid N+1 with Include
var orders = await _db.Orders
    .Include(o => o.Items)
    .ThenInclude(i => i.Product)
    .Where(o => o.UserId == userId)
    .ToListAsync(ct);

// Better: Projection
var orders = await _db.Orders
    .Where(o => o.UserId == userId)
    .Select(o => new OrderDto
    {
        Id = o.Id,
        Total = o.Total,
        Items = o.Items.Select(i => new OrderItemDto
        {
            ProductName = i.Product.Name,
            Quantity = i.Quantity,
        }).ToList(),
    })
    .ToListAsync(ct);
```

---

## 缓存策略

### Redis 查询缓存

```typescript
import Redis from 'ioredis'

const redis = new Redis(process.env.REDIS_URL)

async function cachedQuery<T>(
  key: string,
  queryFn: () => Promise<T>,
  ttlSeconds: number = 300
): Promise<T> {
  const cached = await redis.get(key)
  if (cached) return JSON.parse(cached)

  const result = await queryFn()
  await redis.setex(key, ttlSeconds, JSON.stringify(result))
  return result
}

// Usage
const products = await cachedQuery(
  `products:category:${categoryId}:page:${page}`,
  () => db.product.findMany({ where: { categoryId }, skip, take }),
  300 // 5 minutes
)

// Invalidation
async function invalidateProductCache(categoryId: string) {
  const keys = await redis.keys(`products:category:${categoryId}:*`)
  if (keys.length) await redis.del(...keys)
}
```

### 物化视图

```sql
CREATE MATERIALIZED VIEW monthly_sales AS
SELECT
  DATE_TRUNC('month', created_at) as month,
  category_id,
  COUNT(*) as order_count,
  SUM(total) as revenue,
  AVG(total) as avg_order_value
FROM orders
WHERE created_at >= DATE_TRUNC('year', CURRENT_DATE)
GROUP BY 1, 2;

CREATE UNIQUE INDEX idx_monthly_sales ON monthly_sales(month, category_id);

-- Refresh (can be scheduled via pg_cron)
REFRESH MATERIALIZED VIEW CONCURRENTLY monthly_sales;
```

---

## 连接池配置

### Node.js (pg)

```typescript
import { Pool } from 'pg'

const pool = new Pool({
  max: 20,                      // Max connections
  idleTimeoutMillis: 30000,     // Close idle connections after 30s
  connectionTimeoutMillis: 2000, // Fail fast if can't connect in 2s
  maxUses: 7500,                // Refresh connection after N uses
})

// Monitor pool health
setInterval(() => {
  console.log({
    total: pool.totalCount,
    idle: pool.idleCount,
    waiting: pool.waitingCount,
  })
}, 60000)
```

---

## 监控查询

### 活动连接

```sql
SELECT count(*), state
FROM pg_stat_activity
WHERE datname = current_database()
GROUP BY state;
```

### 长时间运行的查询

```sql
SELECT pid, now() - query_start AS duration, query, state
FROM pg_stat_activity
WHERE (now() - query_start) > interval '5 minutes'
AND state = 'active';
```

### 表大小

```sql
SELECT
  relname AS table,
  pg_size_pretty(pg_total_relation_size(relid)) AS total_size,
  pg_size_pretty(pg_relation_size(relid)) AS data_size,
  pg_size_pretty(pg_total_relation_size(relid) - pg_relation_size(relid)) AS index_size
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC
LIMIT 20;
```

### 表膨胀

```sql
SELECT
  tablename,
  pg_size_pretty(pg_total_relation_size(tablename::regclass)) as size,
  n_dead_tup,
  n_live_tup,
  CASE WHEN n_live_tup > 0
    THEN round(n_dead_tup::numeric / n_live_tup, 2)
    ELSE 0
  END as dead_ratio
FROM pg_stat_user_tables
WHERE n_dead_tup > 1000
ORDER BY dead_ratio DESC;
```

---

## 反模式

1. ❌ `SELECT *` — 始终指定需要的列
2. ❌ 外键缺少索引 — 始终索引 FK 列
3. ❌ `LIKE '%search%'` — 使用全文搜索或三 gram 索引代替
4. ❌ 大型 `IN` 子句 — 使用 `ANY(ARRAY[...])` 或连接值列表
5. ❌ 无界查询缺少 `LIMIT` — 始终分页
6. ❌ 在生产环境中创建索引不使用 `CONCURRENTLY`
7. ❌ 运行迁移前不测试回滚
8. ❌ 忽略 `EXPLAIN ANALYZE` 输出 — 始终验证执行计划
9. ❌ 将金钱存储为 `FLOAT` — 使用 `DECIMAL(10,2)` 或整数分
10. ❌ 缺少 `NOT NULL` 约束 — 明确空值性
