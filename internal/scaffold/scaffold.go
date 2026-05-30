package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Options struct {
	AppName     string
	Module      string
	TargetDir   string
	DefaultPort string
	Force       bool
}

type TemplateData struct {
	AppName      string
	Module       string
	DefaultPort  string
	ProjectTitle string
}

func CreateProject(opts Options) error {
	if strings.TrimSpace(opts.AppName) == "" {
		return errors.New("项目名不能为空")
	}

	if strings.TrimSpace(opts.Module) == "" {
		opts.Module = opts.AppName
	}

	if strings.TrimSpace(opts.DefaultPort) == "" {
		opts.DefaultPort = "8080"
	}

	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = opts.AppName
	}

	info, err := os.Stat(targetDir)
	if err == nil && info.IsDir() && !opts.Force {
		entries, readErr := os.ReadDir(targetDir)
		if readErr != nil {
			return fmt.Errorf("读取目标目录失败：%w", readErr)
		}
		if len(entries) > 0 {
			return fmt.Errorf("目标目录 %q 已存在且不为空，可使用 --force 覆盖", targetDir)
		}
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	data := TemplateData{
		AppName:      opts.AppName,
		Module:       opts.Module,
		DefaultPort:  opts.DefaultPort,
		ProjectTitle: strings.ToUpper(opts.AppName[:1]) + opts.AppName[1:],
	}

	for name, raw := range files() {
		rendered, err := render(raw, data)
		if err != nil {
			return fmt.Errorf("渲染模板 %s 失败：%w", name, err)
		}

		fullPath := filepath.Join(targetDir, name)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("创建子目录失败：%w", err)
		}
		if err := os.WriteFile(fullPath, []byte(rendered), 0o644); err != nil {
			return fmt.Errorf("写入文件 %s 失败：%w", name, err)
		}
	}

	return nil
}

func render(raw string, data TemplateData) (string, error) {
	tmpl, err := template.New("file").Parse(raw)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
