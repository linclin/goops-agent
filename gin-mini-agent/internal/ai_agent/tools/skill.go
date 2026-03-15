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
	"fmt"
	"strings"

	"github.com/cloudwego/eino/adk/middlewares/skill"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type SkillToolImpl struct {
	backend skill.Backend
}

type SkillToolConfig struct {
}

func defaultSkillToolConfig(ctx context.Context) (*SkillToolConfig, error) {
	config := &SkillToolConfig{}
	return config, nil
}

func NewSkillTool(backend skill.Backend) (tool.BaseTool, error) {
	t := &SkillToolImpl{backend: backend}
	tn, err := t.ToEinoTool()
	if err != nil {
		return nil, err
	}
	return tn, nil
}

func (s *SkillToolImpl) ToEinoTool() (tool.InvokableTool, error) {
	return utils.InferTool("skill", "Load and execute a predefined skill", s.Invoke)
}

func (s *SkillToolImpl) Invoke(ctx context.Context, req SkillReq) (res SkillRes, err error) {
	skill, err := s.backend.Get(ctx, req.SkillName)
	if err != nil {
		res.Message = fmt.Sprintf("Failed to get skill: %s", err.Error())
		return res, nil
	}

	skills, _ := s.backend.List(ctx)
	var skillNames []string
	for _, skill := range skills {
		skillNames = append(skillNames, skill.Name)
	}

	res.Message = fmt.Sprintf("Skill loaded successfully. Available skills: %s", strings.Join(skillNames, ", "))
	res.Content = skill.Content
	return res, nil
}

type SkillReq struct {
	SkillName string `json:"skill_name"`
}

type SkillRes struct {
	Message string `json:"message"`
	Content string `json:"content"`
}
