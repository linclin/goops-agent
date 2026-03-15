---
name: ClickHouse
slug: clickhouse
version: 1.0.1
homepage: https://clawic.com/skills/clickhouse
description: 查询、优化和管理 ClickHouse OLAP 数据库，包括模式设计、性能调优和数据导入模式。
metadata: {"clawdbot":{"emoji":"🏠","requires":{"bins":["clickhouse-client"]},"os":["linux","darwin"],"install":[{"id":"brew","kind":"brew","formula":"clickhouse","bins":["clickhouse-client"],"label":"Install ClickHouse (Homebrew)"}]}}
---

# ClickHouse 🏠

对数十亿行数据进行实时分析。亚秒级查询。无需索引。

## 设置

首次使用时，请阅读 `setup.md` 了解连接配置。

## 使用场景

用户需要 OLAP 分析、日志分析、时间序列数据或实时仪表板。Agent 负责处理模式设计、查询优化、数据导入和集群管理。

## 架构

数据存储在 `~/clickhouse/` 目录。参见 `memory-template.md` 了解结构。

```
~/clickhouse/
├── memory.md        # 连接配置 + 查询模式
├── schemas/         # 每个数据库的表定义
└── queries/         # 保存的分析查询
```

## 快速参考

| 主题 | 文件 |
|-------|------|
| 设置和连接 | `setup.md` |
| 内存模板 | `memory-template.md` |
| 查询模式 | `queries.md` |
| 性能调优 | `performance.md` |
| 数据导入 | `ingestion.md` |

## 核心规则

### 1. 始终指定引擎
每个表都需要明确的引擎。默认使用 MergeTree 系列：

```sql
-- 时间序列 / 日志
CREATE TABLE events (
    timestamp DateTime,
    event_type String,
    data String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_type);

-- 聚合指标
CREATE TABLE daily_stats (
    date Date,
    metric String,
    value AggregateFunction(sum, UInt64)
) ENGINE = AggregatingMergeTree()
ORDER BY (date, metric);
```

### 2. ORDER BY 就是你的索引
ClickHouse 没有传统索引。`ORDER BY` 子句决定数据布局：

- 将高基数过滤列放在前面
- 将范围列（日期、时间戳）放在靠前位置
- 匹配你最常用的 WHERE 模式

```sql
-- 好：先按 user_id 过滤，再按日期范围
ORDER BY (user_id, date, event_type)

-- 差：当你要按 user_id 过滤时，日期在前面
ORDER BY (date, user_id, event_type)
```

### 3. 使用适当的数据类型

| 用途 | 类型 | 原因 |
|----------|------|-----|
| 时间戳 | `DateTime` 或 `DateTime64` | 原生时间函数 |
| 低基数字符串 | `LowCardinality(String)` | 10倍压缩 |
| 少量值的枚举 | `Enum8` 或 `Enum16` | 最小存储空间 |
| 仅在需要时使用 Nullable | `Nullable(T)` | 增加开销 |
| IP 地址 | `IPv4` 或 `IPv6` | 4字节 vs 16+字节 |

### 4. 批量插入
永远不要逐行插入。ClickHouse 针对批量写入进行了优化：

```bash
# 好：批量插入
clickhouse-client --query="INSERT INTO events FORMAT JSONEachRow" < batch.json

# 差：循环中逐条插入
for row in data:
    INSERT INTO events VALUES (...)
```

最小批量：1,000 行。最佳：10,000-100,000 行。

### 5. 使用 FINAL 预热查询
在 ReplacingMergeTree/CollapsingMergeTree 上的查询需要 `FINAL` 来保证准确性：

```sql
-- 可能返回重复/旧版本
SELECT * FROM users WHERE id = 123;

-- 保证最新版本
SELECT * FROM users FINAL WHERE id = 123;
```

`FINAL` 有性能开销。对于仪表板，考虑使用物化视图。

### 6. 使用物化视图加速
预聚合昂贵的计算：

```sql
CREATE MATERIALIZED VIEW hourly_events
ENGINE = SummingMergeTree()
ORDER BY (hour, event_type)
AS SELECT
    toStartOfHour(timestamp) AS hour,
    event_type,
    count() AS events
FROM events
GROUP BY hour, event_type;
```

### 7. 首先检查系统表
调试之前，检查系统表：

```sql
-- 正在运行的查询
SELECT * FROM system.processes;

-- 最近的查询性能
SELECT query, elapsed, read_rows, memory_usage
FROM system.query_log
WHERE type = 'QueryFinish'
ORDER BY event_time DESC
LIMIT 10;

-- 表大小
SELECT database, table, formatReadableSize(total_bytes) as size
FROM system.tables
ORDER BY total_bytes DESC;
```

## 常见陷阱

- **使用 String 而非 LowCardinality** → 状态/类型列存储空间大10倍
- **错误的 ORDER BY** → 全表扫描而非索引查找
- **逐行插入** → 大量分区碎片，写入缓慢
- **缺少 TTL** → 表无限增长，磁盘满
- **SELECT *** → 读取所有列，破坏列式存储优势
- **到处使用 Nullable** → 开销 + NULL 处理复杂性
- **忘记 FINAL** → 合并表中出现过期/重复数据

## 性能检查清单

运行昂贵查询之前：

1. **检查 EXPLAIN**：`EXPLAIN SELECT ...` 显示执行计划
2. **先采样**：`SELECT ... FROM table SAMPLE 0.01` 获取1%样本
3. **限制列**：只 SELECT 你需要的
4. **使用 PREWHERE**：在读取所有列之前过滤
5. **检查分区**：`SELECT count() FROM system.parts WHERE table='X'`

```sql
-- PREWHERE 优化
SELECT user_id, event_type, data
FROM events
PREWHERE date = today()
WHERE event_type = 'click';
```

## 集群管理

### 添加 TTL 进行数据保留

```sql
-- 删除旧数据
ALTER TABLE events
MODIFY TTL timestamp + INTERVAL 90 DAY;

-- 移动到冷存储
ALTER TABLE events
MODIFY TTL timestamp + INTERVAL 30 DAY TO VOLUME 'cold';
```

### 监控磁盘使用

```sql
SELECT
    database,
    table,
    formatReadableSize(sum(bytes_on_disk)) as disk_size,
    sum(rows) as total_rows,
    count() as parts
FROM system.parts
WHERE active
GROUP BY database, table
ORDER BY sum(bytes_on_disk) DESC;
```

## 外部端点

| 端点 | 发送数据 | 用途 |
|----------|-----------|---------|
| localhost:8123 | SQL 查询 | HTTP 接口 |
| localhost:9000 | SQL 查询 | 原生 TCP 接口 |

不联系外部服务。所有查询都在用户指定的 ClickHouse 实例上运行。

## 安全与隐私

**本地保存的数据（经用户同意）：**
- 连接配置（主机、端口、数据库）在 ~/clickhouse/memory.md
- 查询模式和模式文档
- 认证方式偏好（密码 vs 证书）

**重要：** 如果你提供数据库密码，它们将以明文形式存储在 ~/clickhouse/ 中。考虑使用环境变量或由 clickhouse-client 管理的连接配置。

**此技能不会：**
- 在没有明确用户配置的情况下连接到任何 ClickHouse
- 发送数据到外部服务
- 自动收集或存储凭据

## 相关技能
如果用户确认，使用 `clawhub install <slug>` 安装：
- `sql` — SQL 查询模式
- `analytics` — 数据分析工作流
- `data-analysis` — 结构化数据探索

## 反馈

- 如果有用：`clawhub star clickhouse`
- 保持更新：`clawhub sync`
