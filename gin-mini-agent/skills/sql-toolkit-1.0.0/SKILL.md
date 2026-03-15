---
name: sql-toolkit
description: 查询、设计、迁移和优化 SQL数据库。用于 SQLite、PostgreSQL 或 MySQL — 架构设计、编写查询、创建迁移、索引、备份/恢复和调试慢查询。不需要 ORM。
metadata: {"clawdbot":{"emoji":"🗄️","requires":{"anyBins":["sqlite3","psql","mysql"]},"os":["linux","darwin","win32"]}}
---

# SQL 工具包

直接从命令行使用关系型数据库。涵盖 SQLite、PostgreSQL 和 MySQL，包括架构设计、查询、迁移、索引和操作的模弌。

## 何时使用

- 创建或修改数据库架构
- 编写复杂查询（连接、聚合、窗口函数、CTE）
- 构建迁移脚本
- 使用索引和 EXPLAIN 优化慢查询
- 备份和恢复数据库
- 使用 SQLite 快速数据探索（零设置）

## SQLite（零设置）

SQLite 随 Python 一起提供，在所有系统上都可用。用于本地数据、原型设计和单文件数据库。

### 快速开始

```bash
# 创建/打开数据库
sqlite3 mydb.sqlite

# 直接导入 CSV
sqlite3 mydb.sqlite ".mode csv" ".import data.csv mytable" "SELECT COUNT(*) FROM mytable;"

# 单行查询
sqlite3 mydb.sqlite "SELECT * FROM users WHERE created_at > '2026-01-01' LIMIT 10;"

# 导出为 CSV
sqlite3 -header -csv mydb.sqlite "SELECT * FROM orders;" > orders.csv

# 带标题和列的交互模式
sqlite3 -header -column mydb.sqlite
```

### 架构操作

```sql
-- 创建表
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

-- 带外键创建
CREATE TABLE orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total REAL NOT NULL CHECK(total >= 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','paid','shipped','cancelled')),
    created_at TEXT DEFAULT (datetime('now'))
);

-- 添加列
ALTER TABLE users ADD COLUMN phone TEXT;

-- 创建索引
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE UNIQUE INDEX idx_users_email ON users(email);

-- 查看架构
.schema users
.tables
```

## PostgreSQL

### 连接

```bash
# Connect
psql -h localhost -U myuser -d mydb

# Connection string
psql "postgresql://user:pass@localhost:5432/mydb?sslmode=require"

# Run single query
psql -h localhost -U myuser -d mydb -c "SELECT NOW();"

# Run SQL file
psql -h localhost -U myuser -d mydb -f migration.sql

# List databases
psql -l
```

### 架构设计模式

```sql
-- 对分布式友好的主键使用 UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('user','admin','moderator')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_unique UNIQUE(email)
);

-- 自动更新 updated_at
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_modtime
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_modified_column();

-- 枚举类型（PostgreSQL 特有）
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'shipped', 'delivered', 'cancelled');

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status order_status NOT NULL DEFAULT 'pending',
    total NUMERIC(10,2) NOT NULL CHECK(total >= 0),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 部分索引（仅索引活动订单 — 更小、更快）
CREATE INDEX idx_orders_active ON orders(user_id, created_at)
    WHERE status NOT IN ('delivered', 'cancelled');

-- GIN 索引用于 JSONB 查询
CREATE INDEX idx_orders_metadata ON orders USING GIN(metadata);
```

### JSONB 查询（PostgreSQL）

```sql
-- 存储 JSON
INSERT INTO orders (user_id, total, metadata)
VALUES ('...', 99.99, '{"source": "web", "coupon": "SAVE10", "items": [{"sku": "A1", "qty": 2}]}');

-- 查询 JSON 字段
SELECT * FROM orders WHERE metadata->>'source' = 'web';
SELECT * FROM orders WHERE metadata->'items' @> '[{"sku": "A1"}]';
SELECT metadata->>'coupon' AS coupon, COUNT(*) FROM orders GROUP BY 1;

-- 更新 JSON 字段
UPDATE orders SET metadata = jsonb_set(metadata, '{source}', '"mobile"') WHERE id = '...';
```

## MySQL

### 连接

```bash
mysql -h localhost -u root -p mydb
mysql -h localhost -u root -p -e "SELECT NOW();" mydb
```

### 与 PostgreSQL 的主要区别

```sql
-- 自增（不是 SERIAL）
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- JSON 类型（MySQL 5.7+）
CREATE TABLE orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    metadata JSON,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 查询 JSON
SELECT * FROM orders WHERE JSON_EXTRACT(metadata, '$.source') = 'web';
-- Or shorthand:
SELECT * FROM orders WHERE metadata->>'$.source' = 'web';
```

## 查询模式

### 连接

```sql
-- 内连接（仅匹配行）
SELECT u.name, o.total, o.status
FROM users u
INNER JOIN orders o ON o.user_id = u.id
WHERE o.created_at > '2026-01-01';

-- 左连接（所有用户，即使没有订单）
SELECT u.name, COUNT(o.id) AS order_count, COALESCE(SUM(o.total), 0) AS total_spent
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.id, u.name;

-- 自连接（查找相同邮件域的用户）
SELECT a.name, b.name, SPLIT_PART(a.email, '@', 2) AS domain
FROM users a
JOIN users b ON SPLIT_PART(a.email, '@', 2) = SPLIT_PART(b.email, '@', 2)
WHERE a.id < b.id;
```

### 聚合

```sql
-- 带 HAVING 的分组
SELECT status, COUNT(*) AS cnt, SUM(total) AS revenue
FROM orders
GROUP BY status
HAVING COUNT(*) > 10
ORDER BY revenue DESC;

-- 运行总计（窗口函数）
SELECT date, revenue,
    SUM(revenue) OVER (ORDER BY date) AS cumulative_revenue
FROM daily_sales;

-- 组内排名
SELECT user_id, total,
    RANK() OVER (PARTITION BY user_id ORDER BY total DESC) AS rank
FROM orders;

-- 移动平均（最后 7 条记录）
SELECT date, revenue,
    AVG(revenue) OVER (ORDER BY date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) AS ma_7
FROM daily_sales;
```

### 公共表表达式（CTE）

```sql
-- 可读的多步骤查询
WITH monthly_revenue AS (
    SELECT DATE_TRUNC('month', created_at) AS month,
           SUM(total) AS revenue
    FROM orders
    WHERE status = 'paid'
    GROUP BY 1
),
growth AS (
    SELECT month, revenue,
           LAG(revenue) OVER (ORDER BY month) AS prev_revenue,
           ROUND((revenue - LAG(revenue) OVER (ORDER BY month)) /
                 NULLIF(LAG(revenue) OVER (ORDER BY month), 0) * 100, 1) AS growth_pct
    FROM monthly_revenue
)
SELECT * FROM growth ORDER BY month;

-- 递归 CTE（组织结构/树遍历）
WITH RECURSIVE org_tree AS (
    SELECT id, name, manager_id, 0 AS depth
    FROM employees
    WHERE manager_id IS NULL
    UNION ALL
    SELECT e.id, e.name, e.manager_id, t.depth + 1
    FROM employees e
    JOIN org_tree t ON e.manager_id = t.id
)
SELECT REPEAT('  ', depth) || name AS org_chart FROM org_tree ORDER BY depth, name;
```

## 迁移

### 手动迁移脚本模式

```bash
#!/bin/bash
# migrate.sh - 运行编号的 SQL 迁移文件
DB_URL="${1:?Usage: migrate.sh <db-url>}"
MIGRATIONS_DIR="./migrations"

# 创建跟踪表
psql "$DB_URL" -c "CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ DEFAULT NOW()
);"

# 按顺序运行待处理的迁移
for file in $(ls "$MIGRATIONS_DIR"/*.sql | sort); do
    version=$(basename "$file" .sql)
    already=$(psql "$DB_URL" -tAc "SELECT 1 FROM schema_migrations WHERE version='$version';")
    if [ "$already" = "1" ]; then
        echo "SKIP: $version (already applied)"
        continue
    fi
    echo "APPLY: $version"
    psql "$DB_URL" -f "$file" && \
    psql "$DB_URL" -c "INSERT INTO schema_migrations (version) VALUES ('$version');" || {
        echo "FAILED: $version"
        exit 1
    }
done
echo "All migrations applied."
```

### 迁移文件约定

```
migrations/
  001_create_users.sql
  002_create_orders.sql
  003_add_users_phone.sql
  004_add_orders_metadata_index.sql
```

每个文件：
```sql
-- 003_add_users_phone.sql
-- Up
ALTER TABLE users ADD COLUMN phone TEXT;

-- To reverse: ALTER TABLE users DROP COLUMN phone;
```

## 查询优化

### EXPLAIN（PostgreSQL）

```sql
-- 显示查询计划
EXPLAIN SELECT * FROM orders WHERE user_id = '...' AND status = 'paid';

-- 显示实际执行时间
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT * FROM orders WHERE user_id = '...' AND status = 'paid';
```

**查找什么：**
- 大表上的 `Seq Scan` → 需要索引
- 大量行的 `Nested Loop` → 考虑 `Hash Join`（可能需要更多 `work_mem`）
- `Rows Removed by Filter` 很高 → 索引不覆盖过滤器
- 实际行数与估计值相差很大 → 运行 `ANALYZE tablename;` 更新统计信息

### 索引策略

```sql
-- 单列（最常见）
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- 复合（用于过滤两列的查询）
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
-- 列顺序很重要：先放等值过滤，最后放范围过滤

-- 覆盖索引（包含数据列以避免表查找）
CREATE INDEX idx_orders_covering ON orders(user_id, status) INCLUDE (total, created_at);

-- 部分索引（更小、更快 — 只索引查询的内容）
CREATE INDEX idx_orders_pending ON orders(user_id) WHERE status = 'pending';

-- 检查未使用的索引
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0 AND indexname NOT LIKE '%pkey%'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### SQLite EXPLAIN

```sql
EXPLAIN QUERY PLAN SELECT * FROM orders WHERE user_id = 5;
-- 查找：SCAN（差）vs SEARCH USING INDEX（好）
```

## 备份和恢复

### PostgreSQL

```bash
# 完整转储（自定义格式，压缩）
pg_dump -Fc -h localhost -U myuser mydb > backup.dump

# 恢复
pg_restore -h localhost -U myuser -d mydb --clean --if-exists backup.dump

# SQL 转储（可移植、可读）
pg_dump -h localhost -U myuser mydb > backup.sql

# 转储特定表
pg_dump -h localhost -U myuser -t users -t orders mydb > partial.sql

# 复制表到 CSV
psql -c "\copy (SELECT * FROM users) TO 'users.csv' CSV HEADER"
```

### SQLite

```bash
# 备份（只需复制文件，但使用 .backup 保证一致性）
sqlite3 mydb.sqlite ".backup backup.sqlite"

# 转储为 SQL
sqlite3 mydb.sqlite .dump > backup.sql

# 从 SQL 恢复
sqlite3 newdb.sqlite < backup.sql
```

### MySQL

```bash
# 转储
mysqldump -h localhost -u root -p mydb > backup.sql

# 恢复
mysql -h localhost -u root -p mydb < backup.sql
```

## 提示

- 在应用程序代码中始终使用参数化查询 — 永远不要将用户输入拼接到 SQL 中
- 在 PostgreSQL 中使用 `TIMESTAMPTZ`（而不是`TIMESTAMP`）用于时区感知日期
- 在 SQLite 中设置 `PRAGMA journal_mode=WAL;` 用于并发读取性能
- 在部署任何运行在大表上的查询前使用 `EXPLAIN`
- PostgreSQL：`\d+ tablename`显示列、索引和大小。`\di+`列出所有索引及其大小
- 快速数据探索，可以将任何 CSV 导入到 SQLite：`sqlite3 :memory: ".mode csv" ".import file.csv t" "SELECT ..."`
