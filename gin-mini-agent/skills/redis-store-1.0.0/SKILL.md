---
name: Redis
description: 有效地使用 Redis 进行缓存、队列和数据结构，包括适当的过期和持久化。
metadata: {"clawdbot":{"emoji":"🔴","requires":{"anyBins":["redis-cli"]},"os":["linux","darwin","win32"]}}
---

## 过期（内存泄漏）

- 没有 TTL 的键永久存在 — 为每个缓存键设置过期时间：`SET key value EX 3600`
- SET 后不能添加 TTL 而不使用另一个命令 — 使用 `SETEX` 或 `SET ... EX`
- `EXPIRE` 在键更新时默认重置 — `SET` 移除 TTL；使用 `SET ... KEEPTTL` (Redis 6+)
- 惰性过期：过期的键在访问时被移除 — 可能在被触碰前消耗内存
- 大数据库使用 `SCAN`：过期的键仍然显示直到清理周期运行

## 我未充分利用的数据结构

- 用于速率限制的有序集合：`ZADD limits:{user} {now} {request_id}` + `ZREMRANGEBYSCORE` 用于滑动窗口
- 用于唯一计数的 HyperLogLog：`PFADD visitors {ip}` 使用 12KB 存储数十亿个唯一值
- 用于队列的 Streams：`XADD`, `XREAD`, `XACK` — 比 LIST 更适合可靠队列
- 用于对象的 Hashes：`HSET user:1 name "Alice" email "a@b.com"` — 比 JSON 字符串更节省内存

## 原子性陷阱

- `GET` 然后 `SET` 不是原子的 — 另一个客户端可以在两者之间修改；使用 `INCR`, `SETNX` 或 Lua
- 使用 `SETNX` 实现锁：`SET lock:resource {token} NX EX 30` — NX = 仅当不存在时
- `WATCH`/`MULTI`/`EXEC` 用于乐观锁定 — 如果监视的键改变则事务中止
- Lua 脚本是原子的 — 用于复杂操作：`EVAL "script" keys args`

## Pub/Sub 限制

- 消息不持久化 — 订阅者在断开连接时错过消息
- 最多一次投递 — 没有确认，没有重试
- 使用 Streams 实现可靠的消息传递 — `XREAD BLOCK` + `XACK` 模式
- 跨集群的 Pub/Sub：消息发送到所有节点 — 可以工作但增加开销

## 持久化配置

- RDB（快照）：快速恢复，但快照之间可能丢失数据 — 默认每 5 分钟
- AOF（追加日志）：数据丢失更少，恢复更慢 — `appendfsync everysec` 是很好的平衡
- 两者都关闭 = 纯缓存 — 如果数据可以重新生成则可以接受
- `BGSAVE` 用于手动快照 — 不阻塞但分叉进程，需要内存余量

## 内存管理（关键）

- `maxmemory` 必须设置 — 没有它，Redis 使用所有 RAM，然后交换 = 灾难
- 淘汰策略：`allkeys-lru` 用于缓存，`volatile-lru` 用于混合，`noeviction` 用于持久数据
- `INFO memory` 显示使用情况 — 监控 `used_memory` vs `maxmemory`
- 大键影响淘汰 — 一个 1GB 的键淘汰效果差；更喜欢许多小键

## 集群

- 哈希槽：键通过哈希分布 — 多键操作需要相同的槽
- 哈希标签：`{user:1}:profile` 和 `{user:1}:sessions` 去同一个槽 — 用于相关键
- 不能跨槽 `MGET`/`MSET` — 除非所有键在同一个槽，否则报错
- `MOVED` 重定向：客户端必须跟随 — 使用支持集群的客户端库

## 常见模式

- 旁路缓存：检查 Redis，未命中 → 获取 DB → 写入 Redis — 标准缓存
- 写穿：同时写入 DB + Redis — 保持缓存新鲜
- 速率限制器：`INCR requests:{ip}:{minute}` 配合 `EXPIRE` — 简单固定窗口
- 分布式锁：`SET ... NX EX` + 唯一 token — 释放时验证 token

## 连接管理

- 连接池：重用连接 — 创建连接很昂贵
- 管道命令：发送批次而不等待 — 减少往返次数
- 关闭时使用 `QUIT` — 优雅断开连接
- Sentinel 或 Cluster 实现高可用 — 单个 Redis 是单点故障

## 常见错误

- 缓存键没有 TTL — 内存增长直到 OOM
- 作为主数据库使用而没有持久化 — 重启后数据丢失
- 在单线程 Redis 中阻塞操作 — `KEYS *` 阻塞一切；使用 `SCAN`
- 存储大型 blob — Redis 是内存；100MB 的值很昂贵
- 忽略 `maxmemory` — 生产环境 Redis 没有限制会搞垮主机
