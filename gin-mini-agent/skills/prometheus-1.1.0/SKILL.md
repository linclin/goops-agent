---
name: prometheus
description: 查询 Prometheus 监控数据以检查服务器指标、资源使用和系统健康状态。当用户询问服务器状态、磁盘空间、CPU/内存使用、网络统计或任何 Prometheus 收集的指标时使用。支持多个 Prometheus 实例的聚合查询、配置文件或环境变量，以及 HTTP Basic Auth。
---

# Prometheus 技能

从一个或多个 Prometheus 实例查询监控数据。支持通过单个命令跨多个 Prometheus 服务器进行联邦查询。

## 快速开始

### 1. 初始设置

运行交互式配置向导：

```bash
cd ~/.openclaw/workspace/skills/prometheus
node scripts/cli.js init
```

这将在你的 OpenClaw 工作区创建一个 `prometheus.json` 配置文件（`~/.openclaw/workspace/prometheus.json`）。

### 2. 开始查询

```bash
# 查询默认实例
node scripts/cli.js query 'up'

# 一次查询所有实例
node scripts/cli.js query 'up' --all

# 列出已配置的实例
node scripts/cli.js instances
```

## 配置

### 配置文件位置

默认情况下，技能在工作区查找配置：

```
~/.openclaw/workspace/prometheus.json
```

**优先级顺序：**
1. `PROMETHEUS_CONFIG` 环境变量指定的路径
2. `~/.openclaw/workspace/prometheus.json`
3. `~/.openclaw/workspace/config/prometheus.json`
4. `./prometheus.json`（当前目录）
5. `~/.config/prometheus/config.json`

### 配置格式

在工作区创建 `prometheus.json`（或使用 `node cli.js init`）：

```json
{
  "instances": [
    {
      "name": "production",
      "url": "https://prometheus.example.com",
      "user": "admin",
      "password": "secret"
    },
    {
      "name": "staging",
      "url": "http://prometheus-staging:9090"
    }
  ],
  "default": "production"
}
```

**字段：**
- `name` — 实例的唯一标识符
- `url` — Prometheus 服务器 URL
- `user` / `password` — 可选的 HTTP Basic Auth 凭据
- `default` — 未指定时使用的实例

### 环境变量（旧方式）

对于单实例设置，可以使用环境变量：

```bash
export PROMETHEUS_URL=https://prometheus.example.com
export PROMETHEUS_USER=admin        # 可选
export PROMETHEUS_PASSWORD=secret   # 可选
```

## 使用方法

### 全局标志

| 标志 | 描述 |
|------|-------------|
| `-c, --config <path>` | 配置文件路径 |
| `-i, --instance <name>` | 目标特定实例 |
| `-a, --all` | 查询所有已配置实例 |

### 命令

#### 设置

```bash
# 交互式配置向导
node scripts/cli.js init
```

#### 查询指标

```bash
cd ~/.openclaw/workspace/skills/prometheus

# 查询默认实例
node scripts/cli.js query 'up'

# 查询特定实例
node scripts/cli.js query 'up' -i staging

# 一次查询所有实例
node scripts/cli.js query 'up' --all

# 自定义配置文件
node scripts/cli.js query 'up' -c /path/to/config.json
```

#### 常用查询

**磁盘空间使用：**
```bash
node scripts/cli.js query '100 - (node_filesystem_avail_bytes / node_filesystem_size_bytes * 100)' --all
```

**CPU 使用率：**
```bash
node scripts/cli.js query '100 - (avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)' --all
```

**内存使用率：**
```bash
node scripts/cli.js query '(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100' --all
```

**负载平均值：**
```bash
node scripts/cli.js query 'node_load1' --all
```

### 列出已配置实例

```bash
node scripts/cli.js instances
```

输出：
```json
{
  "default": "production",
  "instances": [
    { "name": "production", "url": "https://prometheus.example.com", "hasAuth": true },
    { "name": "staging", "url": "http://prometheus-staging:9090", "hasAuth": false }
  ]
}
```

### 其他命令

```bash
# 列出匹配模式的所有指标
node scripts/cli.js metrics 'node_memory_*'

# 获取标签名称
node scripts/cli.js labels --all

# 获取标签值
node scripts/cli.js label-values instance --all

# 查找时间序列
node scripts/cli.js series '{__name__=~"node_cpu_.*", instance=~".*:9100"}' --all

# 获取活动告警
node scripts/cli.js alerts --all

# 获取抓取目标
node scripts/cli.js targets --all
```

## 多实例输出格式

使用 `--all` 时，结果包含所有实例的数据：

```json
{
  "resultType": "vector",
  "results": [
    {
      "instance": "production",
      "status": "success",
      "resultType": "vector",
      "result": [...]
    },
    {
      "instance": "staging",
      "status": "success",
      "resultType": "vector",
      "result": [...]
    }
  ]
}
```

单个实例的错误不会导致整个查询失败 — 它们会在结果数组中以 `"status": "error"` 出现。

## 常用查询参考

| 指标 | PromQL 查询 |
|--------|--------------|
| 磁盘空闲 % | `node_filesystem_avail_bytes / node_filesystem_size_bytes * 100` |
| 磁盘使用 % | `100 - (node_filesystem_avail_bytes / node_filesystem_size_bytes * 100)` |
| CPU 空闲 % | `avg by (instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100` |
| 内存使用 % | `(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100` |
| 网络 RX | `rate(node_network_receive_bytes_total[5m])` |
| 网络 TX | `rate(node_network_transmit_bytes_total[5m])` |
| 运行时间 | `node_time_seconds - node_boot_time_seconds` |
| 服务状态 | `up` |

## 注意事项

- 即时查询的时间范围默认为最近 1 小时
- 使用范围查询 `[5m]` 进行速率计算
- 所有查询返回 JSON，结果在 `data.result` 中
- 实例标签通常显示 `host:port` 格式
- 使用 `--all` 时，查询并行运行以加快结果
- 配置存储在技能目录外，因此在技能更新后仍然保留
