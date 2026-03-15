---
name: docker-essentials
description: Docker 容器管理、镜像操作和调试的基本命令和工作流程。
homepage: https://docs.docker.com/
metadata: {"clawdbot":{"emoji":"🐳","requires":{"bins":["docker"]}}}
---

# Docker 基础

Docker 容器和镜像管理的基本命令。

## 容器生命周期

### 运行容器
```bash
# 从镜像运行容器
docker run nginx

# 后台运行（分离模式）
docker run -d nginx

# 指定名称运行
docker run --name my-nginx -d nginx

# 端口映射运行
docker run -p 8080:80 -d nginx

# 带环境变量运行
docker run -e MY_VAR=value -d app

# 挂载卷运行
docker run -v /host/path:/container/path -d app

# 退出时自动删除
docker run --rm alpine echo "Hello"

# 交互式终端
docker run -it ubuntu bash
```

### 管理容器
```bash
# 列出运行中的容器
docker ps

# 列出所有容器（包括已停止的）
docker ps -a

# 停止容器
docker stop container_name

# 启动已停止的容器
docker start container_name

# 重启容器
docker restart container_name

# 删除容器
docker rm container_name

# 强制删除运行中的容器
docker rm -f container_name

# 删除所有已停止的容器
docker container prune
```

## 容器检查与调试

### 查看日志
```bash
# 显示日志
docker logs container_name

# 跟踪日志（类似 tail -f）
docker logs -f container_name

# 最后100行
docker logs --tail 100 container_name

# 带时间戳的日志
docker logs -t container_name
```

### 执行命令
```bash
# 在运行中的容器内执行命令
docker exec container_name ls -la

# 交互式 shell
docker exec -it container_name bash

# 以特定用户执行
docker exec -u root -it container_name bash

# 带环境变量执行
docker exec -e VAR=value container_name env
```

### 检查
```bash
# 检查容器详情
docker inspect container_name

# 获取特定字段（JSON 路径）
docker inspect -f '{{.NetworkSettings.IPAddress}}' container_name

# 查看容器统计信息
docker stats

# 查看特定容器统计信息
docker stats container_name

# 查看容器内进程
docker top container_name
```

## 镜像管理

### 构建镜像
```bash
# 从 Dockerfile 构建
docker build -t myapp:1.0 .

# 使用自定义 Dockerfile 构建
docker build -f Dockerfile.dev -t myapp:dev .

# 带构建参数构建
docker build --build-arg VERSION=1.0 -t myapp .

# 无缓存构建
docker build --no-cache -t myapp .
```

### 管理镜像
```bash
# 列出镜像
docker images

# 从仓库拉取镜像
docker pull nginx:latest

# 标记镜像
docker tag myapp:1.0 myapp:latest

# 推送到仓库
docker push myrepo/myapp:1.0

# 删除镜像
docker rmi image_name

# 删除未使用的镜像
docker image prune

# 删除所有未使用的镜像
docker image prune -a
```

## Docker Compose

### 基本操作
```bash
# 启动服务
docker-compose up

# 后台启动
docker-compose up -d

# 停止服务
docker-compose down

# 停止并删除卷
docker-compose down -v

# 查看日志
docker-compose logs

# 跟踪特定服务日志
docker-compose logs -f web

# 扩展服务
docker-compose up -d --scale web=3
```

### 服务管理
```bash
# 列出服务
docker-compose ps

# 在服务中执行命令
docker-compose exec web bash

# 重启服务
docker-compose restart web

# 重建服务
docker-compose build web

# 重建并重启
docker-compose up -d --build
```

## 网络

```bash
# 列出网络
docker network ls

# 创建网络
docker network create mynetwork

# 将容器连接到网络
docker network connect mynetwork container_name

# 从网络断开
docker network disconnect mynetwork container_name

# 检查网络
docker network inspect mynetwork

# 删除网络
docker network rm mynetwork
```

## 卷

```bash
# 列出卷
docker volume ls

# 创建卷
docker volume create myvolume

# 检查卷
docker volume inspect myvolume

# 删除卷
docker volume rm myvolume

# 删除未使用的卷
docker volume prune

# 使用卷运行
docker run -v myvolume:/data -d app
```

## 系统管理

```bash
# 查看磁盘使用
docker system df

# 清理所有未使用的资源
docker system prune

# 清理包括未使用的镜像
docker system prune -a

# 清理包括卷
docker system prune --volumes

# 显示 Docker 信息
docker info

# 显示 Docker 版本
docker version
```

## 常用工作流程

**开发容器：**
```bash
docker run -it --rm \
  -v $(pwd):/app \
  -w /app \
  -p 3000:3000 \
  node:18 \
  npm run dev
```

**数据库容器：**
```bash
docker run -d \
  --name postgres \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=mydb \
  -v postgres-data:/var/lib/postgresql/data \
  -p 5432:5432 \
  postgres:15
```

**快速调试：**
```bash
# 进入运行中的容器 shell
docker exec -it container_name sh

# 从容器复制文件
docker cp container_name:/path/to/file ./local/path

# 复制文件到容器
docker cp ./local/file container_name:/path/in/container
```

**多阶段构建：**
```dockerfile
# Dockerfile
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
```

## 常用标志

**`docker run` 标志：**
- `-d`: 分离模式（后台）
- `-it`: 交互式终端
- `-p`: 端口映射（主机:容器）
- `-v`: 卷挂载
- `-e`: 环境变量
- `--name`: 容器名称
- `--rm`: 退出时自动删除
- `--network`: 连接到网络

**`docker exec` 标志：**
- `-it`: 交互式终端
- `-u`: 用户
- `-w`: 工作目录

## 技巧

- 使用 `.dockerignore` 排除构建上下文中的文件
- 在 Dockerfile 中合并 `RUN` 命令以减少层数
- 使用多阶段构建减小镜像大小
- 始终为镜像打上版本标签
- 对一次性容器使用 `--rm`
- 对多容器应用使用 `docker-compose`
- 定期使用 `docker system prune` 清理

## 文档

官方文档：https://docs.docker.com/
Dockerfile 参考：https://docs.docker.com/engine/reference/builder/
Compose 文件参考：https://docs.docker.com/compose/compose-file/
