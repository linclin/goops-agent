---
name: kubectl-skill
description: 通过 kubectl 命令执行和管理 Kubernetes 集群。查询资源、部署应用、调试容器、管理配置和监控集群健康状态。用于处理 Kubernetes 集群、容器、部署或 Pod 诊断时使用。
license: MIT
metadata:
  author: Dennis de Vaal <d.devaal@gmail.com>
  version: "1.0.0"
  keywords: "kubernetes,k8s,container,docker,deployment,pods,cluster"
compatibility: 需要 kubectl 二进制文件（v1.20+）和有效的 kubeconfig 连接到 Kubernetes 集群。支持 macOS、Linux 和 Windows (WSL)。
---

# kubectl 技能

使用 `kubectl` 命令行工具执行 Kubernetes 集群管理操作。

## 概述

此技能使 Agent 能够：
- **查询资源** — 列出和获取 Pod、Deployment、Service、Node 等的详细信息
- **部署和更新** — 创建、应用、修补和更新 Kubernetes 资源
- **调试和故障排除** — 查看日志、在容器中执行命令、检查事件
- **管理配置** — 更新 kubeconfig、切换上下文、管理命名空间
- **监控健康状态** — 检查资源使用、滚动发布状态、事件和 Pod 状态
- **执行操作** — 扩缩容 Deployment、驱逐节点、管理污点和标签

## 前提条件

1. **kubectl 二进制文件** 已安装并可在 PATH 中访问（v1.20+）
2. **kubeconfig** 文件已配置集群凭据（默认：`~/.kube/config`）
3. **有效连接** 到 Kubernetes 集群

## 快速设置

### 安装 kubectl

**macOS:**
```bash
brew install kubernetes-cli
```

**Linux:**
```bash
apt-get install -y kubectl  # Ubuntu/Debian
yum install -y kubectl      # RHEL/CentOS
```

**验证:**
```bash
kubectl version --client
kubectl cluster-info  # 测试连接
```

## 基本命令

### 查询资源
```bash
kubectl get pods                    # 列出当前命名空间的所有 Pod
kubectl get pods -A                 # 所有命名空间
kubectl get pods -o wide            # 显示更多列
kubectl get nodes                   # 列出节点
kubectl describe pod POD_NAME        # 详细信息含事件
```

### 查看日志
```bash
kubectl logs POD_NAME                # 获取日志
kubectl logs -f POD_NAME             # 跟踪日志（tail -f）
kubectl logs POD_NAME -c CONTAINER   # 指定容器
kubectl logs POD_NAME --previous     # 上一个容器的日志
```

### 执行命令
```bash
kubectl exec -it POD_NAME -- /bin/bash   # 交互式 shell
kubectl exec POD_NAME -- COMMAND         # 运行单个命令
```

### 部署应用
```bash
kubectl apply -f deployment.yaml         # 应用配置
kubectl create -f deployment.yaml        # 创建资源
kubectl apply -f deployment.yaml --dry-run=client  # 测试
```

### 更新应用
```bash
kubectl set image deployment/APP IMAGE=IMAGE:TAG  # 更新镜像
kubectl scale deployment/APP --replicas=3          # 扩缩容 Pod
kubectl rollout status deployment/APP              # 检查状态
kubectl rollout undo deployment/APP                # 回滚
```

### 管理配置
```bash
kubectl config view                  # 显示 kubeconfig
kubectl config get-contexts          # 列出上下文
kubectl config use-context CONTEXT   # 切换上下文
```

## 常用模式

### 调试 Pod
```bash
# 1. 识别问题
kubectl describe pod POD_NAME

# 2. 检查日志
kubectl logs POD_NAME
kubectl logs POD_NAME --previous

# 3. 执行调试命令
kubectl exec -it POD_NAME -- /bin/bash

# 4. 检查事件
kubectl get events --sort-by='.lastTimestamp'
```

### 部署新版本
```bash
# 1. 更新镜像
kubectl set image deployment/MY_APP my-app=my-app:v2

# 2. 监控滚动发布
kubectl rollout status deployment/MY_APP -w

# 3. 验证
kubectl get pods -l app=my-app

# 4. 如需回滚
kubectl rollout undo deployment/MY_APP
```

### 准备节点维护
```bash
# 1. 驱逐节点（驱逐所有 Pod）
kubectl drain NODE_NAME --ignore-daemonsets

# 2. 执行维护
# ...

# 3. 恢复上线
kubectl uncordon NODE_NAME
```

## 输出格式

`--output`（`-o`）标志支持多种格式：

- `table` — 默认表格格式
- `wide` — 扩展表格，显示更多列
- `json` — JSON 格式（配合 `jq` 使用）
- `yaml` — YAML 格式
- `jsonpath` — JSONPath 表达式
- `custom-columns` — 定义自定义输出列
- `name` — 仅显示资源名称

**示例：**
```bash
kubectl get pods -o json | jq '.items[0].metadata.name'
kubectl get pods -o jsonpath='{.items[*].metadata.name}'
kubectl get pods -o custom-columns=NAME:.metadata.name,STATUS:.status.phase
```

## 全局标志（适用于所有命令）

```bash
-n, --namespace=<ns>           # 在指定命名空间操作
-A, --all-namespaces           # 跨所有命名空间操作
--context=<context>            # 使用指定的 kubeconfig 上下文
-o, --output=<format>          # 输出格式（json, yaml, table 等）
--dry-run=<mode>               # 试运行模式（none, client, server）
-l, --selector=<labels>        # 按标签过滤
--field-selector=<selector>    # 按字段过滤
-v, --v=<int>                  # 详细程度（0-9）
```

## 试运行模式

- `--dry-run=client` — 快速客户端验证（安全测试命令）
- `--dry-run=server` — 服务端验证（更准确）
- `--dry-run=none` — 实际执行（默认）

**始终先用 `--dry-run=client` 测试：**
```bash
kubectl apply -f manifest.yaml --dry-run=client
```

## 高级主题

详细参考材料、逐命令文档、故障排除指南和高级工作流程，请参阅：
- [references/REFERENCE.md](references/REFERENCE.md) — 完整 kubectl 命令参考
- [scripts/](scripts/) — 常用任务的辅助脚本

## 实用技巧

1. **使用标签选择器进行批量操作：**
   ```bash
   kubectl delete pods -l app=myapp
   kubectl get pods -l env=prod,tier=backend
   ```

2. **实时监控资源：**
   ```bash
   kubectl get pods -w  # 监控变化
   ```

3. **使用 `-A` 标志查看所有命名空间：**
   ```bash
   kubectl get pods -A  # 查看所有 Pod
   ```

4. **保存输出以便后续比较：**
   ```bash
   kubectl get deployment my-app -o yaml > deployment-backup.yaml
   ```

5. **删除前先检查：**
   ```bash
   kubectl delete pod POD_NAME --dry-run=client
   ```

## 获取帮助

```bash
kubectl help                      # 通用帮助
kubectl COMMAND --help            # 命令帮助
kubectl explain pods              # 资源文档
kubectl explain pods.spec         # 字段文档
```

## 环境变量

- `KUBECONFIG` — kubeconfig 文件路径（可包含多个路径，用 `:` 分隔）
- `KUBECTL_CONTEXT` — 覆盖默认上下文

## 资源

- [官方 kubectl 文档](https://kubernetes.io/docs/reference/kubectl/)
- [kubectl 速查表](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Kubernetes API 参考](https://kubernetes.io/docs/reference/generated/kubernetes-api/)
- [Agent 技能规范](https://agentskills.io/)

---

**版本：** 1.0.0  
**许可证：** MIT  
**兼容：** kubectl v1.20+, Kubernetes v1.20+
