---
name: Ansible
description: 避免常见的 Ansible 错误 — YAML 语法陷阱、变量优先级、幂等性失败和处理器问题。
metadata: {"clawdbot":{"emoji":"🔧","requires":{"bins":["ansible"]},"os":["linux","darwin"]}}
---

## YAML 语法陷阱
- Jinja2 值需要引号 — `"{{ variable }}"` 而不是 `{{ variable }}`
- 字符串中的 `:` 需要引号 — `msg: "Note: this works"` 而不是 `msg: Note: this`
- 布尔字符串：`yes`, `no`, `true`, `false` 会被解析为布尔值 — 字面字符串需要引号
- 缩进必须一致 — 标准 2 个空格，禁止使用制表符

## 变量优先级
- Extra vars (`-e`) 优先级最高 — 覆盖所有其他变量
- 主机变量优于组变量 — 更具体的获胜
- playbook 中的 `vars:` 优于 inventory 变量 — 顺序：inventory < playbook < extra vars
- 未定义变量会失败 — 使用 `{{ var | default('fallback') }}`

## 幂等性
- `command`/`shell` 模块不是幂等的 — 总是显示"changed"，使用 `creates:` 或特定模块
- 使用 `apt`, `yum`, `copy` 等 — 专为幂等性设计
- 对于不改变状态的命令使用 `changed_when: false` — 如查询操作
- 使用 `creates:`/`removes:` 实现命令幂等性 — 文件存在/不存在时跳过

## 处理器（Handlers）
- Handlers 仅在任务报告 changed 时运行 — 不在"ok"时运行
- Handlers 在 play 结束时运行一次 — 不是在 notify 后立即运行
- 多次 notify 同一个处理器 = 只运行一次 — 去重
- 使用 `--force-handlers` 即使在失败时也运行 — 或使用 `meta: flush_handlers`

## Become（权限提升）
- 使用 `become: yes` 以 root 身份运行 — 使用 `become_user:` 指定特定用户
- `become_method: sudo` 是默认值 — 如需使用 `su` 或 `doas`
- sudo 需要密码 — 使用 `--ask-become-pass` 或在 ansible.cfg 中配置
- 某些模块需要在任务级别设置 become — 即使 playbook 已有 `become: yes`

## 条件语句
- `when:` 不使用 Jinja2 括号 — `when: ansible_os_family == "Debian"` 而不是 `when: "{{ ... }}"`
- 多个条件使用 `and`/`or` — 或使用列表表示隐式 `and`
- 使用 `is defined`, `is not defined` 检查可选变量 — `when: my_var is defined`
- 布尔变量：`when: my_bool` — 不要比较 `== true`

## 循环
- `loop:` 是现代方式，`with_items:` 是旧方式 — 两者都可用，推荐 loop
- 嵌套循环使用 `loop_control.loop_var` — 避免变量冲突
- `item` 是循环变量 — 使用 `loop_control.label` 获得更清晰的输出
- 使用 `until:` 进行重试循环 — `until: result.rc == 0 retries: 5 delay: 10`

## Facts（事实）
- `gather_facts: no` 加速 play — 但不能使用 `ansible_*` 变量
- Facts 通过 `fact_caching` 缓存 — 在多次运行间持久化
- 自定义 facts 放在 `/etc/ansible/facts.d/*.fact` — JSON 或 INI 格式，作为 `ansible_local` 可用

## 常见错误
- `register:` 即使失败也会捕获输出 — 检查 `result.rc` 或 `result.failed`
- `ignore_errors: yes` 继续执行但不改变结果 — 任务在 register 中仍显示"failed"
- `delegate_to: localhost` 用于本地命令 — 但 `local_action` 更简洁
- 加密文件需要 Vault 密码 — 使用 `--ask-vault-pass` 或 vault 密码文件
- `--check`（干跑模式）不被所有模块支持 — `command`, `shell` 总是跳过
