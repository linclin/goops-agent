/*
 * Copyright 2025 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package ai_agent 提供 AI Agent 的核心功能实现
//
// 该包基于 CloudWeGo Eino 框架构建，实现了一个完整的 AI 对话代理系统。
// 主要功能包括：
//   - 知识库检索增强生成（RAG）
//   - 对话历史管理与检索
//   - 多工具调用（网络搜索、文件操作等）
//   - 流式响应输出
//
// 架构说明：
//
//	Graph 执行流程:
//	START -> InputToQuery -> Retriever (知识库检索) -----> ChatTemplate -> ReactAgent -> END
//	                      -> ConversationRetriever (对话历史检索) ->
//	START -> InputToHistory ------------------------------------------------->
//
// 节点说明：
//   - InputToQuery: 将用户消息转换为查询字符串
//   - InputToHistory: 将用户消息转换为模板变量
//   - Retriever: 从知识库检索相关文档
//   - ConversationRetriever: 从向量数据库检索对话历史
//   - ChatTemplate: 组装提示词模板
//   - ReactAgent: 执行推理和工具调用
package ai_agent

import (
	"context"
	"fmt"
	"gin-mini-agent/pkg/global"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// GlobalAgent 全局 AI Agent 实例
//
// 该实例在应用启动时通过 BuildAiAgent 函数初始化，
// 是一个可运行的 Graph 对象，支持 Invoke（同步调用）和 Stream（流式调用）两种模式。
//
// 使用示例:
//
//	// 同步调用
//	resp, err := ai_agent.GlobalAgent.Invoke(ctx, userMessage)
//
//	// 流式调用
//	streamReader, err := ai_agent.GlobalAgent.Stream(ctx, userMessage)
var GlobalAgent compose.Runnable[*UserMessage, *schema.Message]

// GlobalConversationManager 全局对话历史管理器
//
// 负责对话历史的存储和检索，支持多种向量数据库后端：
//   - Chromem: 本地文件存储，适合开发和小规模部署
//   - Redis: 分布式存储，适合中等规模部署
//   - Milvus: 分布式向量数据库，适合大规模部署
//
// 对话历史存储格式:
//
//	用户：{用户问题}
//	助手：{AI 回复}
//
// 使用示例:
//
//	// 存储对话
//	err := ai_agent.GlobalConversationManager.Store(ctx, userQuery, aiResponse)
//
//	// 检索对话历史
//	docs, err := ai_agent.GlobalConversationManager.Retrieve(ctx, query)
var GlobalConversationManager *ConversationManager

// BuildAiAgent 构建 AI Agent 图结构
//
// 该函数是 AI Agent 的核心构建函数，负责创建和配置整个处理流程图。
// 图结构采用有向无环图（DAG）设计，支持并行执行和条件分支。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - skillBackend: 可选的 Skill 后端，用于添加 skill 工具
//   - skillMiddleware: 可选的 Skill 中间件，用于提取技能指令
//
// 返回:
//   - r: 可运行的 Graph 实例
//   - err: 构建过程中的错误
//
// 图结构说明:
//
//	节点列表:
//	- InputToQuery: Lambda 节点，提取用户查询字符串
//	- InputToHistory: Lambda 节点，准备模板变量
//	- Retriever: 检索器节点，从知识库检索相关文档
//	- ConversationRetriever: 检索器节点，从向量数据库检索对话历史
//	- ChatTemplate: 模板节点，组装系统提示词
//	- ReactAgent: Lambda 节点，执行推理和工具调用
//
//	边列表:
//	- START -> InputToQuery: 开始节点连接到查询转换节点
//	- START -> InputToHistory: 开始节点连接到历史处理节点
//	- InputToQuery -> Retriever: 查询转换后进行知识库检索
//	- InputToQuery -> ConversationRetriever: 查询转换后进行对话历史检索
//	- Retriever -> ChatTemplate: 知识库检索结果传入模板
//	- ConversationRetriever -> ChatTemplate: 对话历史检索结果传入模板
//	- InputToHistory -> ChatTemplate: 模板变量传入模板
//	- ChatTemplate -> ReactAgent: 模板渲染后执行推理
//	- ReactAgent -> END: 推理完成后结束
//
// 执行模式:
//   - WithNodeTriggerMode(compose.AllPredecessor): 所有前驱节点完成后才触发当前节点
//   - 这意味着 ChatTemplate 需要等待 Retriever、ConversationRetriever 和 InputToHistory 全部完成
func BuildAiAgent(ctx context.Context, skillBackend skill.Backend, skillMiddleware adk.AgentMiddleware) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	// 定义节点名称常量，便于维护和理解
	const (
		// InputToQuery 节点名称：将 UserMessage 转换为查询字符串
		InputToQuery = "InputToQuery"
		// ChatTemplate 节点名称：组装提示词模板
		ChatTemplate = "ChatTemplate"
		// ReactAgent 节点名称：执行推理和工具调用
		ReactAgent = "ReactAgent"
		// Retriever 节点名称：知识库检索器
		Retriever = "Retriever"
		// ConversationRetriever 节点名称：对话历史检索器
		ConversationRetriever = "ConversationRetriever"
		// InputToHistory 节点名称：将 UserMessage 转换为模板变量
		InputToHistory = "InputToHistory"
	)

	// 创建新的 Graph 实例
	// 输入类型: *UserMessage (用户消息)
	// 输出类型: *schema.Message (AI 回复)
	g := compose.NewGraph[*UserMessage, *schema.Message]()

	// 添加 InputToQuery Lambda 节点
	// 功能: 从 UserMessage 中提取 Query 字段作为查询字符串
	// 该查询字符串将用于知识库检索和对话历史检索
	_ = g.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(inputToQuery), compose.WithNodeName("UserMessageToQuery"))

	// 添加 ChatTemplate 模板节点
	// 功能: 根据系统提示词、知识库文档、对话历史和用户输入组装最终提示词
	// 如果提供了 Skill Middleware，提取技能指令并注入到系统提示词中
	skillInstruction := ""
	if skillMiddleware.AdditionalInstruction != "" {
		skillInstruction = skillMiddleware.AdditionalInstruction
	}
	chatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx, skillInstruction)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplateKeyOfChatTemplate)

	// 添加 ReactAgent Lambda 节点
	// 功能：执行 ReAct 推理循环，支持工具调用
	// ReactAgent 会根据提示词决定是否调用工具，直到生成最终答案
	reactAgentKeyOfLambda, err := newReactAgent(ctx, skillMiddleware)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(ReactAgent, reactAgentKeyOfLambda, compose.WithNodeName("ReAct Agent"))

	// 添加知识库检索器节点
	// 功能: 根据用户查询从知识库中检索相关文档
	// 支持多种向量数据库后端: Chromem, Redis, Milvus
	// 检索结果将作为 "documents" 键传入模板
	retrieverKeyOfRetriever, err := newRetriever(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(Retriever, retrieverKeyOfRetriever, compose.WithOutputKey("documents"))

	// 创建嵌入模型实例
	// 嵌入模型用于将文本转换为向量，用于向量检索
	embedder, err := newEmbedding(ctx)
	if err != nil {
		return nil, err
	}

	// 创建对话历史管理器
	// 对话历史管理器负责存储和检索对话历史
	// 使用独立的向量集合存储对话历史，不影响知识库数据
	conversationManager, err := NewConversationManager(ctx, embedder)
	if err != nil {
		return nil, err
	}
	GlobalConversationManager = conversationManager

	// 添加对话历史检索器节点
	// 功能: 根据用户查询从向量数据库中检索相关的历史对话
	// 检索结果将作为 "conversation_history" 键传入模板
	_ = g.AddRetrieverNode(ConversationRetriever, conversationManager.retriever, compose.WithOutputKey("conversation_history"))

	// 添加 InputToHistory Lambda 节点
	// 功能: 将 UserMessage 转换为模板变量
	// 输出包含: content (用户查询), history (前端传入的历史), date (当前时间)
	_ = g.AddLambdaNode(InputToHistory, compose.InvokableLambdaWithOption(inputToHistory), compose.WithNodeName("UserMessageToVariables"))

	// ==================== 构建图边 ====================

	// 开始节点 -> InputToQuery
	// 用户消息首先进入查询转换节点
	_ = g.AddEdge(compose.START, InputToQuery)

	// 开始节点 -> InputToHistory
	// 用户消息同时进入历史处理节点（并行执行）
	_ = g.AddEdge(compose.START, InputToHistory)

	// InputToQuery -> Retriever
	// 查询字符串传入知识库检索器
	_ = g.AddEdge(InputToQuery, Retriever)

	// InputToQuery -> ConversationRetriever
	// 查询字符串同时传入对话历史检索器（并行执行）
	_ = g.AddEdge(InputToQuery, ConversationRetriever)

	// Retriever -> ChatTemplate
	// 知识库检索结果传入模板节点
	_ = g.AddEdge(Retriever, ChatTemplate)

	// ConversationRetriever -> ChatTemplate
	// 对话历史检索结果传入模板节点
	_ = g.AddEdge(ConversationRetriever, ChatTemplate)

	// InputToHistory -> ChatTemplate
	// 模板变量传入模板节点
	_ = g.AddEdge(InputToHistory, ChatTemplate)

	// ChatTemplate -> ReactAgent
	// 组装好的提示词传入推理节点
	_ = g.AddEdge(ChatTemplate, ReactAgent)

	// ReactAgent -> END
	// 推理完成后结束
	_ = g.AddEdge(ReactAgent, compose.END)
	// 编译图
	// WithGraphName: 设置图的名称，便于调试和监控
	// WithNodeTriggerMode: 设置节点触发模式为 AllPredecessor
	//   即所有前驱节点完成后才触发当前节点执行
	r, err = g.Compile(ctx, compose.WithGraphName("AiAgent"),
		compose.WithNodeTriggerMode(compose.AllPredecessor),
	)
	if err != nil {
		return nil, err
	}
	graphHandler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			global.Log.Info(fmt.Sprintf("[Graph Start] component=%s name=%s input=%T", info.Component, info.Name, input))
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			global.Log.Info(fmt.Sprintf("[Graph End] component=%s name=%s output=%T", info.Component, info.Name, output))
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			global.Log.Info(fmt.Sprintf("[Graph Error] component=%s name=%s err=%v", info.Component, info.Name, err))
			return ctx
		}).
		Build()

	// Register as global callbacks (applies to all subsequent runs)
	callbacks.AppendGlobalHandlers(graphHandler)
	return r, err
}
