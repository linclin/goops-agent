<h1 align="center">gin-mini-agent</h1>
 

## 主要功能
- 最小化Gin脚手架，适合Agent开发场景 

## 使用golang大众开源类库 
- [gin](https://github.com/gin-gonic/gin) 一款高效的golang web框架 [教程](https://gin-gonic.com/zh-cn/docs/) 
- [viper](https://github.com/spf13/viper)  配置管理工具, 支持多种配置文件类型.[教程](https://darjun.github.io/2020/01/18/godailylib/viper/)
- [lumberjack](https://github.com/natefinch/lumberjack) 日志切割工具, 高效分离大日志文件, 按日期保存文件
- [cast](https://github.com/spf13/cast) 一个小巧、实用的类型转换库，用于将一个类型转为另一个类型 [教程](https://darjun.github.io/2020/01/20/godailylib/cast/)  
- [cron](https://github.com/robfig/cron) 实现了 cron 规范解析器和任务运行器，简单来讲就是包含了定时任务所需的功能  [教程](https://darjun.github.io/2020/06/25/godailylib/cron) 
- [lo](https://github.com/samber/lo) 基于泛型实现的Golang工具库 

## API 接口文档

### Agent 聊天接口

#### 1. SSE 流式聊天接口

**请求信息**
- 路径：`POST /api/v1/agent/chat`
- 认证：Basic Auth
- Content-Type：`application/json`

**请求参数**
```json
{
  "id": "会话ID",
  "query": "用户问题",
  "history": [
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ]
}
```

**请求示例**
```bash
curl -X POST http://localhost:8080/api/v1/agent/chat \
  -H "Content-Type: application/json" \
  -u "username:password" \
  -d '{
    "id": "session-001",
    "query": "你好，请介绍一下自己",
    "history": []
  }'
```

**响应格式（SSE）**
```
event: start
data: 

data: {"event":"message","data":"你"}
data: {"event":"message","data":"好"}
data: {"event":"message","data":"，"}
...
event: done
data: 
```

**事件类型说明**
- `start`：开始生成响应
- `message`：生成的文本内容（逐字符推送）
- `error`：发生错误
- `done`：生成完成

#### 2. 非流式聊天接口

**请求信息**
- 路径：`POST /api/v1/agent/chat/non-stream`
- 认证：Basic Auth
- Content-Type：`application/json`

**请求参数**
```json
{
  "id": "会话ID",
  "query": "用户问题",
  "history": [
    {"role": "user", "content": "..."},
    {"role": "assistant", "content": "..."}
  ]
}
```

**请求示例**
```bash
curl -X POST http://localhost:8080/api/v1/agent/chat/non-stream \
  -H "Content-Type: application/json" \
  -u "username:password" \
  -d '{
    "id": "session-001",
    "query": "你好",
    "history": []
  }'
```

**响应示例**
```json
{
  "code": 200,
  "data": {
    "id": "session-001",
    "content": "你好！我是 AI 助手...",
    "role": "assistant"
  },
  "message": "成功"
}
```

### RAG 索引接口

#### 执行 RAG 索引

**请求信息**
- 路径：`POST /api/v1/rag/index`
- 认证：Basic Auth
- Content-Type：`application/json`

**请求参数**
```json
{
  "dir": "要索引的目录路径",
  "dbType": "向量数据库类型（可选：redis/chromem/milvus，默认 redis）"
}
```

**请求示例**
```bash
curl -X POST http://localhost:8080/api/v1/rag/index \
  -H "Content-Type: application/json" \
  -u "username:password" \
  -d '{
    "dir": "./documents",
    "dbType": "chromem"
  }'
```

**响应示例**
```json
{
  "code": 200,
  "data": {
    "message": "索引成功",
    "dir": "./documents",
    "dbType": "chromem"
  },
  "message": "成功"
}
```
