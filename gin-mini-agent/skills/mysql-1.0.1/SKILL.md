---
name: MySQL
slug: mysql
version: 1.0.1
description: 编写正确的 MySQL 查询，包括适当的字符集、索引、事务和生产模式。
metadata: {"clawdbot":{"emoji":"🐬","requires":{"bins":["mysql"]},"os":["linux","darwin","win32"]}}
---

## 快速参考

| 主题 | 文件 |
|-------|------|
| 索引设计深入 | `indexes.md` |
| 事务和锁定 | `transactions.md` |
| 查询优化 | `queries.md` |
| 生产配置 | `production.md` | |

## 字符集陷阱

- `utf8` 已损坏 — 只有 3 字节，无法存储 emoji；始终使用 `utf8mb4`
- `utf8mb4_unicode_ci` 用于不区分大小写排序；`utf8mb4_bin` 用于精确字节比较
- JOIN 中的排序规则冲突会破坏性能 — 确保表之间排序规则一致
- 连接字符集必须匹配：`SET NAMES utf8mb4` 或连接字符串参数
- utf8mb4 列上的索引更大 — 可能达到索引大小限制；考虑前缀索引

## 与 PostgreSQL 的索引差异

- 没有部分索引 — 不能在索引定义中使用 `WHERE active = true`
- 在 MySQL 8.0.13 之前没有表达式索引 — 在此之前必须使用生成列
- TEXT/BLOB 需要前缀长度：`INDEX (description(100))` — 没有长度会报错
- 没有 INCLUDE 用于覆盖索引 — 将列添加到索引本身：`INDEX (a, b, c)` 来覆盖 c
- 外键仅在 InnoDB 中自动索引 — 假设之前验证引擎

## UPSERT 模式

- `INSERT ... ON DUPLICATE KEY UPDATE` — 不是标准 SQL；需要唯一键冲突
- `LAST_INSERT_ID()` 用于自增 — 没有像 PostgreSQL 那样的 RETURNING 子句
- `REPLACE INTO` 先删除再插入 — 改变自增 ID，触发 DELETE 级联
- 检查受影响的行数：1 = 已插入，2 = 已更新（反直觉）

## 锁定陷阱

- `SELECT ... FOR UPDATE` 锁定行 — 但间隙锁定可能锁定比预期更多的行
- InnoDB 使用 next-key 锁定 — 防止幻读但可能导致死锁
- 锁定等待超时默认 50 秒 — 使用 `innodb_lock_wait_timeout` 调整
- `FOR UPDATE SKIP LOCKED` 存在于 MySQL 8+ — 队列模式
- InnoDB 默认隔离级别是 REPEATABLE READ，而不是 PostgreSQL 的 READ COMMITTED
- 死锁是预期的 — 代码必须捕获并重试，而不仅仅是失败

## GROUP BY 严格性

- `sql_mode` 默认包含 `ONLY_FULL_GROUP_BY` 在 MySQL 5.7+
- 非聚合列必须在 GROUP BY 中 — 不像旧 MySQL 的宽松模式
- `ANY_VALUE(column)` 用于静音错误，当你知道值相同时
- 在遗留数据库上检查 sql_mode — 可能行为不同

## InnoDB vs MyISAM

- 始终使用 InnoDB — 事务、行锁定、外键、崩溃恢复
- MyISAM 仍然是某些系统表的默认值 — 不要用于应用程序数据
- 检查引擎：`SHOW TABLE STATUS` — 使用 `ALTER TABLE ... ENGINE=InnoDB` 转换
- JOIN 中混合引擎可以工作但失去事务保证

## 查询怪癖

- `LIMIT offset, count` 顺序与 PostgreSQL 的 `LIMIT count OFFSET offset` 不同
- `!=` 和 `<>` 都可以用；推荐使用 `<>` 符合 SQL 标准
- 非事务性 DDL — `ALTER TABLE` 立即提交，不能回滚
- 布尔类型是 `TINYINT(1)` — `TRUE`/`FALSE` 只是 1/0
- `IFNULL(a, b)` 用于两个参数而不是 `COALESCE` — 虽然 COALESCE 也可以用

## 连接管理

- `wait_timeout` 终止空闲连接 — 默认 8 小时；连接池可能注意不到
- `max_connections` 默认 151 — 通常太低；每个连接使用内存
- 连接池：不要超过所有应用实例的 max_connections 总和
- `SHOW PROCESSLIST` 查看活动连接 — 使用 `KILL <id>`` 终止长时间运行的连接

## 复制感知

- 基于语句的复制可能因非确定性函数而损坏 — UUID()、NOW()
- 基于行的复制更安全但带宽更多 — MySQL 8 默认
- 只读副本有延迟 — 依赖副本读取前检查 `Seconds_Behind_Master`
- 不要写入副本 — 通常是只读的但要验证

## 性能

- `EXPLAIN ANALYZE` 仅在 MySQL 8.0.18+ — 旧版本只有 EXPLAIN 没有实际时间
- 查询缓存在 MySQL 8 中已移除 — 不要依赖它；在应用程序级别缓存
- `OPTIMIZE TABLE` 用于碎片化的表 — 锁定表；对于大表使用 pt-online-schema-change
- `innodb_buffer_pool_size` — 专用数据库服务器设置为 RAM 的 70-80%
