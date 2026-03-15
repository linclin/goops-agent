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

// Package tools 提供 AI Agent 可调用的工具集合
//
// 该包定义了多种工具，扩展 AI Agent 的能力边界。
// 工具是 Agent 与外部世界交互的桥梁，允许 Agent 执行文件操作、
// 浏览器自动化、网络请求等任务。
//
// 当前可用工具:
//   - kubectl: Kubernetes 集群管理工具
//
// 工具开发指南:
//  1. 定义工具结构体，包含配置信息
//  2. 实现 ToEinoTool 方法，返回 tool.InvokableTool
//  3. 实现 Invoke 方法，执行具体的工具逻辑
//  4. 使用 utils.InferTool 自动生成工具信息
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sSchema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubectlToolImpl kubectl 工具实现
//
// 该工具用于管理 Kubernetes 集群，支持 get、describe、create、delete、apply 等操作。
// 支持多集群管理，通过 ~/.kube/ 目录下的集群名称文件加载配置。
//
// 使用场景:
//   - 用户要求查看 Kubernetes 集群中的资源
//   - 用户要求创建、删除或更新 Kubernetes 资源
//   - 用户要求管理多个 Kubernetes 集群
//
// 示例:
//   - 获取 pods: kubectl get pods
//   - 创建 deployment: kubectl create deployment nginx --image=nginx
//   - 应用配置: kubectl apply -f deployment.yaml
//   - 多集群操作: kubectl get pods --cluster=ST1-MAIN-A1

type KubectlToolImpl struct {
	// config 工具配置
	config *KubectlToolConfig
	// kubeClient kubernetes 客户端
	kubeClient *kubernetes.Clientset
	// dynamicClient dynamic 客户端
	dynamicClient dynamic.Interface
	// restConfig rest 配置
	restConfig *rest.Config
	// namespace 默认命名空间
	namespace string
}

// KubectlToolConfig kubectl 工具配置
//
// 定义了 kubectl 工具的配置选项，包括 kubeconfig 文件路径、集群名称等。
type KubectlToolConfig struct {
	// KubeConfigPath kubeconfig 文件路径
	// 如果为空，会尝试使用默认路径 (~/.kube/config) 或 ~/.kube/ 目录下的文件
	KubeConfigPath string `json:"kube_config_path" jsonschema_description:"kubeconfig 文件路径"`

	// ContextName 要使用的 kubeconfig 上下文
	// 如果为空，会使用 kubeconfig 中的当前上下文
	ContextName string `json:"context_name" jsonschema_description:"要使用的 kubeconfig 上下文"`

	// ClusterName 要使用的集群名称
	// 如果为空，会尝试使用默认集群
	ClusterName string `json:"cluster_name" jsonschema_description:"要使用的集群名称"`

	// KubeConfigDir kubeconfig 目录路径
	// 如果为空，会使用默认路径 (~/.kube/)
	KubeConfigDir string `json:"kube_config_dir" jsonschema_description:"kubeconfig 目录路径"`
}

// defaultKubectlToolConfig 创建默认配置
//
// 返回默认的工具配置实例。
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - *KubectlToolConfig: 配置实例
//   - error: 错误信息
func defaultKubectlToolConfig() *KubectlToolConfig {
	return &KubectlToolConfig{
		KubeConfigPath: "",
		ContextName:    "",
		ClusterName:    "",
		KubeConfigDir:  "",
	}
}

// NewKubectlTool 创建 kubectl 工具实例
//
// 该函数创建一个用于管理 Kubernetes 集群的工具。
// 如果未提供配置，将使用默认配置。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - config: 工具配置，可选
//
// 返回:
//   - tool.BaseTool: 工具实例
//   - error: 创建过程中的错误
//
// 使用示例:
//
//	tool, err := NewKubectlTool(ctx, nil)
//	result, err := tool.Invoke(ctx, KubectlReq{Command: "get", Resource: "pods"})
func NewKubectlTool(ctx context.Context, config *KubectlToolConfig) (tn tool.BaseTool, err error) {
	slog.InfoContext(ctx, "[kubectl] 创建 kubectl 工具")

	// 如果配置为空，使用默认配置
	if config == nil {
		config = defaultKubectlToolConfig()
	}

	// 构建 rest.Config
	restConfig, err := buildRestConfig(config)
	if err != nil {
		slog.ErrorContext(ctx, "[kubectl] 构建 rest.Config 失败", "error", err)
		return nil, fmt.Errorf("构建 rest.Config 失败: %w", err)
	}

	// 创建 kubernetes 客户端
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		slog.ErrorContext(ctx, "[kubectl] 创建 kubernetes 客户端失败", "error", err)
		return nil, fmt.Errorf("创建 kubernetes 客户端失败: %w", err)
	}

	// 创建 dynamic 客户端
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		slog.ErrorContext(ctx, "[kubectl] 创建 dynamic 客户端失败", "error", err)
		return nil, fmt.Errorf("创建 dynamic 客户端失败: %w", err)
	}

	// 获取默认命名空间
	namespace := getDefaultNamespace()

	slog.InfoContext(ctx, "[kubectl] kubectl 工具创建成功", "namespace", namespace)

	// 创建工具实例
	t := &KubectlToolImpl{
		config:        config,
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		restConfig:    restConfig,
		namespace:     namespace,
	}

	// 转换为 Eino 工具
	tn, err = t.ToEinoTool()
	if err != nil {
		return nil, err
	}
	return tn, nil
}

// ToEinoTool 转换为 Eino 工具接口
//
// 该方法将工具实现转换为 Eino 框架的工具接口。
// 使用 utils.InferTool 自动推断工具的参数和返回值类型。
//
// 返回:
//   - tool.InvokableTool: Eino 工具实例
//   - error: 转换错误
func (k *KubectlToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	// 使用 InferTool 自动生成工具信息
	// 参数:
	//   - name: 工具名称，用于 Agent 识别
	//   - description: 工具描述，帮助 Agent 理解工具用途
	//   - invoke: 工具调用函数
	return utils.InferTool("kubectl", "Kubernetes 集群管理工具，支持 get、describe、create、delete、apply 等操作，支持多集群管理", k.Invoke)
}

// KubectlReq kubectl 请求结构体
//
// 定义了 kubectl 工具的输入参数。
type KubectlReq struct {
	// Command 要执行的命令
	// 支持的命令: get, describe, create, delete, apply
	Command string `json:"command" jsonschema_description:"要执行的命令，支持 get、describe、create、delete、apply"`

	// Resource 资源类型
	// 例如: pod, deployment, service 等
	Resource string `json:"resource" jsonschema_description:"资源类型，例如 pod、deployment、service 等"`

	// Name 资源名称
	// 对于 get、describe、delete 命令，指定资源名称
	Name string `json:"name" jsonschema_description:"资源名称，对于 get、describe、delete 命令需要指定"`

	// Namespace 命名空间
	// 如果为空，使用默认命名空间
	Namespace string `json:"namespace" jsonschema_description:"命名空间，如果为空，使用默认命名空间"`

	// Content 资源内容
	// 对于 create、apply 命令，提供 YAML 或 JSON 格式的资源内容
	Content string `json:"content" jsonschema_description:"资源内容，对于 create、apply 命令需要提供 YAML 或 JSON 格式的内容"`

	// Cluster 集群名称
	// 如果为空，使用默认集群
	Cluster string `json:"cluster" jsonschema_description:"集群名称，如果为空，使用默认集群"`

	// Flags 命令参数
	// 例如: {"all-namespaces": true}
	Flags map[string]interface{} `json:"flags" jsonschema_description:"命令参数，例如 {\"all-namespaces\": true}"`
}

// KubectlRes kubectl 响应结构体
//
// 定义了 kubectl 工具的输出结果。
type KubectlRes struct {
	// Result 执行结果
	// 包含命令执行的详细输出
	Result string `json:"result" jsonschema_description:"执行结果，包含命令执行的详细输出"`

	// Success 是否执行成功
	// true 表示执行成功，false 表示执行失败
	Success bool `json:"success" jsonschema_description:"是否执行成功，true 表示执行成功，false 表示执行失败"`

	// Namespace 使用的命名空间
	// 执行命令时使用的命名空间
	Namespace string `json:"namespace" jsonschema_description:"使用的命名空间，执行命令时使用的命名空间"`

	// Cluster 使用的集群
	// 执行命令时使用的集群
	Cluster string `json:"cluster" jsonschema_description:"使用的集群，执行命令时使用的集群"`
}

// Invoke 执行 kubectl 命令
//
// 该方法是工具的核心实现，负责执行指定的 kubectl 命令。
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - req: kubectl 请求，包含命令、资源类型、名称等信息
//
// 返回:
//   - KubectlRes: 执行结果，包含执行状态、结果内容等
//   - error: 执行错误
//
// 支持的命令:
//   - get: 获取资源信息
//   - describe: 获取资源详细信息
//   - create: 创建新资源
//   - delete: 删除资源
//   - apply: 应用资源配置
func (k *KubectlToolImpl) Invoke(ctx context.Context, req KubectlReq) (res KubectlRes, err error) {
	slog.InfoContext(ctx, "[kubectl] 调用 kubectl 工具", "request", req)

	// 验证命令参数
	if req.Command == "" {
		slog.WarnContext(ctx, "[kubectl] 缺少命令参数")
		res.Result = "缺少命令参数"
		res.Success = false
		res.Namespace = k.namespace
		res.Cluster = k.config.ClusterName
		return res, nil
	}

	// 检查客户端是否有效
	if k.kubeClient == nil || k.dynamicClient == nil {
		slog.WarnContext(ctx, "[kubectl] Kubernetes 客户端未初始化")
		res.Result = "Kubernetes 客户端未初始化，请确保已配置 kubeconfig 文件或在集群内运行"
		res.Success = false
		res.Namespace = k.namespace
		res.Cluster = k.config.ClusterName
		return res, nil
	}

	// 检查配置是否有效
	if k.restConfig == nil || k.restConfig.Host == "" || k.restConfig.Host == "https://localhost:8443" {
		slog.WarnContext(ctx, "[kubectl] Kubernetes 配置无效")
		res.Result = "Kubernetes 配置无效，请确保已配置 kubeconfig 文件或在集群内运行"
		res.Success = false
		res.Namespace = k.namespace
		res.Cluster = k.config.ClusterName
		return res, nil
	}

	// 处理命名空间
	namespace := req.Namespace
	if namespace == "" {
		namespace = k.namespace
	}

	// 处理集群
	cluster := req.Cluster
	if cluster == "" {
		cluster = k.config.ClusterName
	}

	// 执行命令
	var result string
	switch req.Command {
	case "get":
		if req.Resource == "" {
			res.Result = "get 命令需要指定资源类型"
			res.Success = false
			res.Namespace = namespace
			res.Cluster = cluster
			return res, nil
		}
		result, err = k.handleGet(ctx, req.Resource, req.Name, namespace, req.Flags)
	case "describe":
		if req.Resource == "" || req.Name == "" {
			res.Result = "describe 命令需要指定资源类型和名称"
			res.Success = false
			res.Namespace = namespace
			res.Cluster = cluster
			return res, nil
		}
		result, err = k.handleDescribe(ctx, req.Resource, req.Name, namespace)
	case "create":
		if req.Resource == "" || req.Content == "" {
			res.Result = "create 命令需要指定资源类型和内容"
			res.Success = false
			res.Namespace = namespace
			res.Cluster = cluster
			return res, nil
		}
		result, err = k.handleCreate(ctx, req.Resource, namespace, req.Content)
	case "delete":
		if req.Resource == "" || req.Name == "" {
			res.Result = "delete 命令需要指定资源类型和名称"
			res.Success = false
			res.Namespace = namespace
			res.Cluster = cluster
			return res, nil
		}
		result, err = k.handleDelete(ctx, req.Resource, req.Name, namespace)
	case "apply":
		if req.Content == "" {
			res.Result = "apply 命令需要提供资源内容"
			res.Success = false
			res.Namespace = namespace
			res.Cluster = cluster
			return res, nil
		}
		result, err = k.handleApply(ctx, namespace, req.Content)
	default:
		res.Result = fmt.Sprintf("不支持的命令: %s", req.Command)
		res.Success = false
		res.Namespace = namespace
		res.Cluster = cluster
		return res, nil
	}

	// 处理执行结果
	if err != nil {
		slog.ErrorContext(ctx, "[kubectl] 命令执行失败", "error", err)
		res.Result = err.Error()
		res.Success = false
	} else {
		slog.InfoContext(ctx, "[kubectl] 命令执行成功")
		res.Result = result
		res.Success = true
	}

	res.Namespace = namespace
	res.Cluster = cluster
	return res, nil
}

// buildRestConfig 构建 rest.Config
//
// 根据配置构建 Kubernetes 客户端的 rest.Config。
// 支持从默认路径或指定路径加载 kubeconfig 文件。
// 如果没有找到 kubeconfig 文件且无法使用集群内配置，返回一个默认配置，以便工具能够初始化。
//
// 参数:
//   - config: 工具配置
//
// 返回:
//   - *rest.Config: rest 配置
//   - error: 构建错误
func buildRestConfig(config *KubectlToolConfig) (*rest.Config, error) {
	// 确定 kubeconfig 目录
	kubeConfigDir := config.KubeConfigDir
	if kubeConfigDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// 获取用户主目录失败，使用当前目录
			kubeConfigDir = "."
		} else {
			kubeConfigDir = filepath.Join(homeDir, ".kube")
		}
	}

	// 确定 kubeconfig 文件路径
	kubeConfigPath := config.KubeConfigPath
	if kubeConfigPath == "" {
		// 如果指定了集群名称，尝试使用集群名称作为文件名
		if config.ClusterName != "" {
			kubeConfigPath = filepath.Join(kubeConfigDir, config.ClusterName)
			// 检查文件是否存在
			if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
				// 文件不存在，使用默认路径
				kubeConfigPath = filepath.Join(kubeConfigDir, "config")
			}
		} else {
			// 使用默认路径
			kubeConfigPath = filepath.Join(kubeConfigDir, "config")
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
		// 文件不存在，尝试使用集群内配置
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			// 无法使用集群内配置，返回一个默认配置
			// 这样工具可以初始化，但在实际调用时会失败
			slog.Warn("无法加载 kubeconfig 文件，也无法使用集群内配置，将使用默认配置")
			return &rest.Config{
				Host: "https://localhost:8443",
			}, nil
		}
		return restConfig, nil
	}

	// 加载 kubeconfig 文件
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: kubeConfigPath,
	}

	// 构建配置
	configOverrides := &clientcmd.ConfigOverrides{}
	if config.ContextName != "" {
		configOverrides.CurrentContext = config.ContextName
	}

	// 创建客户端配置
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, configOverrides,
	)

	// 获取 rest.Config
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		// 加载 kubeconfig 文件失败，返回一个默认配置
		slog.Warn("加载 kubeconfig 文件失败，将使用默认配置", "error", err)
		return &rest.Config{
			Host: "https://localhost:8443",
		}, nil
	}

	return restConfig, nil
}

// getDefaultNamespace 获取默认命名空间
//
// 返回默认的 Kubernetes 命名空间。
//
// 返回:
//   - string: 默认命名空间
func getDefaultNamespace() string {
	// 尝试从环境变量获取
	namespace := os.Getenv("KUBECONFIG_NAMESPACE")
	if namespace != "" {
		return namespace
	}

	// 默认为 default
	return "default"
}

// handleGet 处理 get 命令
//
// 执行 kubectl get 命令，获取指定资源的信息。
//
// 参数:
//   - ctx: 上下文
//   - resource: 资源类型
//   - name: 资源名称
//   - namespace: 命名空间
//   - flags: 命令参数
//
// 返回:
//   - string: 执行结果
//   - error: 执行错误
func (k *KubectlToolImpl) handleGet(ctx context.Context, resource, name, namespace string, flags map[string]interface{}) (string, error) {
	// 处理 --all-namespaces 标志
	if flags != nil {
		if val, ok := flags["all-namespaces"].(bool); ok && val {
			namespace = ""
		}
	}

	// 构建 GVR
	gvr, err := getGVR(resource)
	if err != nil {
		return "", err
	}

	// 创建 dynamic 客户端
	dynamicClient := k.dynamicClient

	// 获取资源
	if name != "" {
		// 获取单个资源
		resourceInterface := dynamicClient.Resource(gvr).Namespace(namespace)
		obj, err := resourceInterface.Get(ctx, name, v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return "", fmt.Errorf("资源 %s/%s 不存在", resource, name)
			}
			return "", err
		}

		// 转换为 JSON
		data, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return "", err
		}

		return string(data), nil
	} else {
		// 获取资源列表
		resourceInterface := dynamicClient.Resource(gvr).Namespace(namespace)
		list, err := resourceInterface.List(ctx, v1.ListOptions{})
		if err != nil {
			return "", err
		}

		// 转换为 JSON
		data, err := json.MarshalIndent(list, "", "  ")
		if err != nil {
			return "", err
		}

		return string(data), nil
	}
}

// handleDescribe 处理 describe 命令
//
// 执行 kubectl describe 命令，获取指定资源的详细信息。
//
// 参数:
//   - ctx: 上下文
//   - resource: 资源类型
//   - name: 资源名称
//   - namespace: 命名空间
//
// 返回:
//   - string: 执行结果
//   - error: 执行错误
func (k *KubectlToolImpl) handleDescribe(ctx context.Context, resource, name, namespace string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("describe 命令需要指定资源名称")
	}

	// 构建 GVR
	gvr, err := getGVR(resource)
	if err != nil {
		return "", err
	}

	// 获取资源
	dynamicClient := k.dynamicClient
	resourceInterface := dynamicClient.Resource(gvr).Namespace(namespace)
	obj, err := resourceInterface.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("资源 %s/%s 不存在", resource, name)
		}
		return "", err
	}

	// 转换为 JSON
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// handleCreate 处理 create 命令
//
// 执行 kubectl create 命令，创建新的资源。
//
// 参数:
//   - ctx: 上下文
//   - resource: 资源类型
//   - namespace: 命名空间
//   - content: 资源内容
//
// 返回:
//   - string: 执行结果
//   - error: 执行错误
func (k *KubectlToolImpl) handleCreate(ctx context.Context, resource, namespace, content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("create 命令需要提供资源内容")
	}

	// 构建 GVR
	gvr, err := getGVR(resource)
	if err != nil {
		return "", err
	}

	// 解析 YAML 或 JSON 内容
	obj, err := parseContent(content)
	if err != nil {
		return "", err
	}

	// 创建资源
	dynamicClient := k.dynamicClient
	resourceInterface := dynamicClient.Resource(gvr).Namespace(namespace)
	createdObj, err := resourceInterface.Create(ctx, obj, v1.CreateOptions{})
	if err != nil {
		return "", err
	}

	// 转换为 JSON
	data, err := json.MarshalIndent(createdObj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// handleDelete 处理 delete 命令
//
// 执行 kubectl delete 命令，删除指定的资源。
//
// 参数:
//   - ctx: 上下文
//   - resource: 资源类型
//   - name: 资源名称
//   - namespace: 命名空间
//
// 返回:
//   - string: 执行结果
//   - error: 执行错误
func (k *KubectlToolImpl) handleDelete(ctx context.Context, resource, name, namespace string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("delete 命令需要指定资源名称")
	}

	// 构建 GVR
	gvr, err := getGVR(resource)
	if err != nil {
		return "", err
	}

	// 删除资源
	dynamicClient := k.dynamicClient
	resourceInterface := dynamicClient.Resource(gvr).Namespace(namespace)
	err = resourceInterface.Delete(ctx, name, v1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("资源 %s/%s 不存在", resource, name)
		}
		return "", err
	}

	return fmt.Sprintf("资源 %s/%s 删除成功", resource, name), nil
}

// handleApply 处理 apply 命令
//
// 执行 kubectl apply 命令，应用资源配置。
//
// 参数:
//   - ctx: 上下文
//   - namespace: 命名空间
//   - content: 资源内容
//
// 返回:
//   - string: 执行结果
//   - error: 执行错误
func (k *KubectlToolImpl) handleApply(ctx context.Context, namespace, content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("apply 命令需要提供资源内容")
	}

	// 解析 YAML 或 JSON 内容
	obj, err := parseContent(content)
	if err != nil {
		return "", err
	}

	// 获取资源类型和名称
	resource := obj.GetKind()
	name := obj.GetName()

	if resource == "" {
		return "", fmt.Errorf("资源内容中缺少 kind 字段")
	}

	if name == "" {
		return "", fmt.Errorf("资源内容中缺少 name 字段")
	}

	// 构建 GVR
	gvr, err := getGVR(resource)
	if err != nil {
		return "", err
	}

	// 应用资源
	dynamicClient := k.dynamicClient
	resourceInterface := dynamicClient.Resource(gvr).Namespace(namespace)

	// 尝试更新资源
	updatedObj, err := resourceInterface.Update(ctx, obj, v1.UpdateOptions{})
	if err != nil {
		// 如果资源不存在，尝试创建
		if errors.IsNotFound(err) {
			createdObj, err := resourceInterface.Create(ctx, obj, v1.CreateOptions{})
			if err != nil {
				return "", err
			}
			data, err := json.MarshalIndent(createdObj, "", "  ")
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
		return "", err
	}

	// 转换为 JSON
	data, err := json.MarshalIndent(updatedObj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// getGVR 获取 GVR (GroupVersionResource)
//
// 根据资源类型获取对应的 GroupVersionResource。
//
// 参数:
//   - resource: 资源类型
//
// 返回:
//   - k8sSchema.GroupVersionResource: GVR
//   - error: 错误信息
func getGVR(resource string) (k8sSchema.GroupVersionResource, error) {
	// 常见资源的 GVR 映射
	gvrMap := map[string]k8sSchema.GroupVersionResource{
		"pod":                    {Group: "", Version: "v1", Resource: "pods"},
		"pods":                   {Group: "", Version: "v1", Resource: "pods"},
		"deployment":             {Group: "apps", Version: "v1", Resource: "deployments"},
		"deployments":            {Group: "apps", Version: "v1", Resource: "deployments"},
		"service":                {Group: "", Version: "v1", Resource: "services"},
		"services":               {Group: "", Version: "v1", Resource: "services"},
		"configmap":              {Group: "", Version: "v1", Resource: "configmaps"},
		"configmaps":             {Group: "", Version: "v1", Resource: "configmaps"},
		"secret":                 {Group: "", Version: "v1", Resource: "secrets"},
		"secrets":                {Group: "", Version: "v1", Resource: "secrets"},
		"namespace":              {Group: "", Version: "v1", Resource: "namespaces"},
		"namespaces":             {Group: "", Version: "v1", Resource: "namespaces"},
		"node":                   {Group: "", Version: "v1", Resource: "nodes"},
		"nodes":                  {Group: "", Version: "v1", Resource: "nodes"},
		"persistentvolume":       {Group: "", Version: "v1", Resource: "persistentvolumes"},
		"persistentvolumes":      {Group: "", Version: "v1", Resource: "persistentvolumes"},
		"persistentvolumeclaim":  {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		"persistentvolumeclaims": {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		"statefulset":            {Group: "apps", Version: "v1", Resource: "statefulsets"},
		"statefulsets":           {Group: "apps", Version: "v1", Resource: "statefulsets"},
		"daemonset":              {Group: "apps", Version: "v1", Resource: "daemonsets"},
		"daemonsets":             {Group: "apps", Version: "v1", Resource: "daemonsets"},
		"replicaset":             {Group: "apps", Version: "v1", Resource: "replicasets"},
		"replicasets":            {Group: "apps", Version: "v1", Resource: "replicasets"},
		"ingress":                {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
		"ingresses":              {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
		"serviceaccount":         {Group: "", Version: "v1", Resource: "serviceaccounts"},
		"serviceaccounts":        {Group: "", Version: "v1", Resource: "serviceaccounts"},
		"role":                   {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		"roles":                  {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		"rolebinding":            {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		"rolebindings":           {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		"clusterrole":            {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		"clusterroles":           {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		"clusterrolebinding":     {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
		"clusterrolebindings":    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
	}

	if gvr, ok := gvrMap[resource]; ok {
		return gvr, nil
	}

	return k8sSchema.GroupVersionResource{}, fmt.Errorf("不支持的资源类型: %s", resource)
}

// parseContent 解析 YAML 或 JSON 内容
//
// 解析 YAML 或 JSON 格式的资源内容。
//
// 参数:
//   - content: 资源内容
//
// 返回:
//   - *unstructured.Unstructured: 解析后的资源对象
//   - error: 解析错误
func parseContent(content string) (*unstructured.Unstructured, error) {
	// 尝试解析 YAML
	obj := &unstructured.Unstructured{}
	err := yaml.Unmarshal([]byte(content), obj)
	if err == nil {
		return obj, nil
	}

	// 尝试解析 JSON
	err = json.Unmarshal([]byte(content), obj)
	if err == nil {
		return obj, nil
	}

	return nil, fmt.Errorf("无法解析资源内容: %w", err)
}
