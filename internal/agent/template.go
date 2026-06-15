package agent

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed prompts/*.prompt
var promptFS embed.FS

// PromptManager 管理 LLM Prompt 模板
type PromptManager struct {
	cache map[string]*template.Template
}

// NewPromptManager 创建 Prompt 管理器并预加载所有模板
func NewPromptManager() (*PromptManager, error) {
	pm := &PromptManager{
		cache: make(map[string]*template.Template),
	}

	entries, err := promptFS.ReadDir("prompts")
	if err != nil {
		return nil, fmt.Errorf("read prompts dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".prompt") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".prompt")
		content, err := promptFS.ReadFile("prompts/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read prompt %s: %w", entry.Name(), err)
		}

		tmpl, err := template.New(name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("parse prompt %s: %w", entry.Name(), err)
		}

		pm.cache[name] = tmpl
	}

	return pm, nil
}

// Render 渲染指定名称的 Prompt 模板
func (pm *PromptManager) Render(name string, data interface{}) (string, error) {
	tmpl, ok := pm.cache[name]
	if !ok {
		return "", fmt.Errorf("prompt not found: %s", name)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render prompt %s: %w", name, err)
	}

	return buf.String(), nil
}
