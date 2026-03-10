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

package ai_agent

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// GlobalAgent 全局 AI Agent 实例
var GlobalAgent compose.Runnable[*UserMessage, *schema.Message]

// GlobalConversationManager 全局对话历史管理器
var GlobalConversationManager *ConversationManager

func BuildAiAgent(ctx context.Context) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	const (
		InputToQuery          = "InputToQuery"
		ChatTemplate          = "ChatTemplate"
		ReactAgent            = "ReactAgent"
		Retriever             = "Retriever"
		ConversationRetriever = "ConversationRetriever"
		InputToHistory        = "InputToHistory"
	)
	g := compose.NewGraph[*UserMessage, *schema.Message]()

	_ = g.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(inputToQuery), compose.WithNodeName("UserMessageToQuery"))

	chatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplateKeyOfChatTemplate)

	reactAgentKeyOfLambda, err := newReactAgent(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(ReactAgent, reactAgentKeyOfLambda, compose.WithNodeName("ReAct Agent"))

	// 知识库检索器
	retrieverKeyOfRetriever, err := newRetriever(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddRetrieverNode(Retriever, retrieverKeyOfRetriever, compose.WithOutputKey("documents"))

	// 对话历史管理器
	embedder, err := newEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	conversationManager, err := NewConversationManager(ctx, embedder)
	if err != nil {
		return nil, err
	}
	GlobalConversationManager = conversationManager

	// 对话历史检索器
	_ = g.AddRetrieverNode(ConversationRetriever, conversationManager.retriever, compose.WithOutputKey("conversation_history"))

	_ = g.AddLambdaNode(InputToHistory, compose.InvokableLambdaWithOption(inputToHistory), compose.WithNodeName("UserMessageToVariables"))

	// 构建图结构
	_ = g.AddEdge(compose.START, InputToQuery)
	_ = g.AddEdge(compose.START, InputToHistory)
	_ = g.AddEdge(InputToQuery, Retriever)
	_ = g.AddEdge(InputToQuery, ConversationRetriever)
	_ = g.AddEdge(Retriever, ChatTemplate)
	_ = g.AddEdge(ConversationRetriever, ChatTemplate)
	_ = g.AddEdge(InputToHistory, ChatTemplate)
	_ = g.AddEdge(ChatTemplate, ReactAgent)
	_ = g.AddEdge(ReactAgent, compose.END)

	r, err = g.Compile(ctx, compose.WithGraphName("AiAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
