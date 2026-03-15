---
name: Linux
description: 操作 Linux 系统，避免权限陷阱、静默失败和常见管理错误。
metadata: {"clawdbot":{"emoji":"🐧","os":["linux","darwin"]}}
---

# Linux 陷阱

## 权限陷阱
- `chmod 777` 解决不了问题，只会搞砸一切 — 找出实际的属主/属组问题
- 脚本上的 setuid 会被安全忽略 — 只对二进制文件有效
- `chown -R` 会跟随目标目录外的符号链接 — 使用 `--no-dereference`
- 默认 umask 022 使文件全局可读 — 敏感系统应设置 077
- ACL 会静默覆盖传统权限 — 用 `getfacl` 检查

## 进程陷阱
- `kill` 默认发送 SIGTERM，不是 SIGKILL — 进程可以忽略它
- `nohup` 对已运行的进程无效 — 改用 `disown`
- 使用 `&` 的后台作业在终端关闭时仍会死亡，除非使用 `disown` 或 `nohup`
- 僵尸进程无法被杀死 — 父进程必须调用 wait() 或被杀死
- `kill -9` 跳过清理处理器 — 可能导致数据丢失，先用 SIGTERM

## 文件系统陷阱
- 删除已打开的文件不会释放空间，直到进程关闭它 — 用 `lsof +L1` 检查
- `rm -rf /path /` 意外加空格 = 灾难 — 使用 `rm -rf /path/` 末尾斜杠
- inode 耗尽但磁盘显示有空闲空间 — 大量小文件问题
- 符号链接循环导致无限递归 — `find -L` 会跟随它们
- `/tmp` 在重启时清空 — 不要在那里存储持久数据

## 磁盘空间谜题
- 被进程打开的已删除文件 — `lsof +L1` 显示它们，重启进程释放空间
- 保留块（默认 5%）仅限 root — `tune2fs -m 1` 可减少
- 日志占用空间 — `journalctl --vacuum-size=500M`
- Docker overlay 占用空间 — `docker system prune -a`
- 快照占用空间 — 检查 LVM、ZFS 或云提供商快照

## 网络
- `localhost` 和 `127.0.0.1` 可能解析不同 — 检查 `/etc/hosts`
- 防火墙规则在重启后清空，除非保存 — `iptables-save` 或使用 firewalld/ufw 持久化
- `netstat` 已弃用 — 改用 `ss`
- 1024 以下端口需要 root — 使用 `setcaps` 获取能力替代
- 高负载下 TCP TIME_WAIT 耗尽 — 调整 `net.ipv4.tcp_tw_reuse`

## SSH 陷阱
- ~/.ssh 权限错误 = 静默认证失败 — 目录 700，密钥 600
- Agent 转发会将你的密钥暴露给远程管理员 — 不信任的服务器上避免使用
- 服务器重建后 known hosts 哈希不匹配 — 用 `ssh-keygen -R` 删除旧条目
- SSH 配置 Host 块：首个匹配生效 — 将特定主机放在通配符之前
- 空闲时连接超时 — 在配置中添加 `ServerAliveInterval 60`

## Systemd
- `systemctl enable` 不会启动服务 — 还需要 `start`
- `restart` vs `reload`：restart 断开连接，reload 不会（如果支持）
- 默认重启后日志丢失 — 在 journald.conf 中设置 `Storage=persistent`
- 失败的服务默认不重试 — 在单元中添加 `Restart=on-failure`
- 网络依赖：`After=network.target` 不够 — 使用 `network-online.target`

## Cron 陷阱
- Cron 只有最小 PATH — 使用绝对路径或在 crontab 中设置 PATH
- 输出默认发邮件 — 重定向到文件或 `/dev/null`
- Cron 使用系统时区，不是用户的 — 如需要，在 crontab 中设置 TZ
- 编辑错误会丢失 crontab — 编辑前先 `crontab -l > backup`
- @reboot 在守护进程重启时也会运行，不只是系统重启

## 内存和 OOM
- OOM killer 选择"最佳"受害者，通常不是肇事者 — 检查 dmesg 中的杀死记录
- 交换区抖动比 OOM 更糟 — 用 `vmstat` 监控
- `free` 中的内存使用包括缓存 — "available" 才是关键
- 进程内存在 `/proc/[pid]/status` — VmRSS 是实际使用量
- cgroups 限制在系统 OOM 之前生效 — 容器先死

## 会撒谎的命令
- `df` 显示文件系统容量，不是物理磁盘 — 检查底层设备
- `du` 不能正确计算稀疏文件 — 文件显示比磁盘使用量小
- `ps aux` 内存百分比可能超过 100%（共享内存被多次计算）
- `uptime` 负载平均值包括不可中断 I/O 等待 — 不只是 CPU
- `top` CPU 百分比是每核心 — 400% 意味着 4 个核心满载
