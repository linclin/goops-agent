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

package tools

import (
	"context"

	"github.com/cloudwego/eino-ext/components/tool/commandline"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type StrReplaceEditorToolImpl struct {
	config *commandline.EditorConfig
}

func NewFileEditorTool(ctx context.Context, config *commandline.EditorConfig) (tool.BaseTool, error) {
	if config == nil {
		config = &commandline.EditorConfig{}
	}
	t := &StrReplaceEditorToolImpl{config: config}
	return t.ToEinoTool()
}

func (s *StrReplaceEditorToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("str_replace_editor", "文件编辑器，支持创建、查看、编辑文件，使用字符串替换方式修改文件内容", s.Invoke)
}

func (s *StrReplaceEditorToolImpl) Invoke(ctx context.Context, req commandline.StrReplaceEditorParams) (string, error) {
	sre, err := commandline.NewStrReplaceEditor(ctx, &commandline.EditorConfig{
		Operator: s.config.Operator,
	})
	if err != nil {
		return "", err
	}
	_, err = sre.Info(ctx)
	if err != nil {
		return "", err
	}
	result, err := sre.Execute(ctx, &req)
	if err != nil {
		return result, err
	}
	return result, nil
}
