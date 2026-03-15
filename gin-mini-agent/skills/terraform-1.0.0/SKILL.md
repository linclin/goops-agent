---
name: Terraform
description: 避免常见的 Terraform 错误 — 状态损坏、count vs for_each、生命周期陷阱和依赖顺序。
metadata: {"clawdbot":{"emoji":"🟪","requires":{"bins":["terraform"]},"os":["linux","darwin","win32"]}}
---

## 状态管理
- 本地状态会损坏/丢失 — 使用远程后端（S3、GCS、Terraform Cloud）
- 多人同时运行 — 使用 DynamoDB 或等效工具启用状态锁定
- 永远不要手动编辑状态 — 使用 `terraform state mv`、`rm`、`import`
- 状态包含明文秘密 — 在休息时加密，限制访问

## Count vs for_each
- `count` 使用索引 — 移除第 0 项会偏移所有索引，强制重新创建
- `for_each` 使用键 — 稳定，移除一个不影响其他
- 不能在同一资源上使用两者 — 选一个
- `for_each` 需要集合或映射 — 使用 `toset()` 转换列表

## 生命周期规则
- `prevent_destroy = true` — 阻止意外删除，必须移除才能销毁
- `create_before_destroy = true` — 新资源在旧资源销毁前创建，用于零停机
- `ignore_changes` 用于外部修改 — `ignore_changes = [tags]` 忽略漂移
- `replace_triggered_by` 强制重新创建 — 当依赖改变时

## 依赖
- 通过引用隐式依赖 — `aws_instance.foo.id` 创建自动依赖
- `depends_on` 用于隐藏依赖 — 当配置中没有引用时
- `depends_on` 接受列表 — `depends_on = [aws_iam_role.x, aws_iam_policy.y]`
- 数据源在计划期间运行 — 如果资源不存在可能失败

## 数据源
- 数据源读取现有资源 — 不创建
- 在计划时运行 — 依赖必须在计划前存在
- 如果隐式依赖不清楚使用 `depends_on` — 否则计划失败
- 考虑使用资源输出代替 — 更明确

## 模块
- 固定模块版本 — `source = "org/name/aws?version=1.2.3"`
- `terraform init -upgrade` 更新 — 不会自动更新
- 模块输出必须显式定义 — 不能从外部访问内部资源
- 嵌套模块：输出必须逐层冒泡 — 每层都需要导出

## 变量
- 无类型 = any — 显式 `type = string`, `list(string)`, `map(object({...}))`
- `sensitive = true` 隐藏输出 — 但状态文件中仍然是明文
- `validation` 块用于约束 — 自定义错误消息
- `nullable = false` 拒绝 null — 默认是可空的

## 常见错误
- `terraform destroy` 是永久的 — 没有撤销，小心使用 `-target`
- 计划成功 ≠ 应用成功 — API 错误、配额、权限在应用时发现
- 重命名资源 = 删除 + 创建 — 使用 `moved` 块或 `terraform state mv`
- Workspaces 不用于环境 — 每个环境使用独立的状态文件/后端
- Provisioners 是最后手段 — 使用 cloud-init、user_data 或配置管理代替

## 导入
- `terraform import aws_instance.foo i-1234` — 将现有资源导入状态
- 不生成配置 — 必须手动编写匹配的资源块
- `import` 块（TF 1.5+） — 配置中的声明式导入
- 导入后计划验证 — 如果配置匹配应该显示无变化
