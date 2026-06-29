package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func runCreateModule(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("请提供模块名称，例如：Golider create-module my-module")
	}

	moduleName := strings.TrimSpace(args[0])
	if moduleName == "" {
		return fmt.Errorf("模块名称不能为空")
	}

	if err := validateModuleName(moduleName); err != nil {
		return err
	}

	targetDir := "internal/addon/modules"
	if len(args) > 1 {
		targetDir = filepath.Clean(args[1])
	}

	fileName := moduleNameToFileName(moduleName)
	filePath := filepath.Join(targetDir, fileName+".go")

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("模块文件 %q 已存在", filePath)
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	content := generateModuleSkeleton(moduleName)

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入模块文件失败：%w", err)
	}

	terminalUI := newTerminalUI(os.Stdout)
	terminalUI.Header("Golider 自定义模块创建")
	terminalUI.KeyValue("模块名称", moduleName)
	terminalUI.KeyValue("生成文件", filePath)
	terminalUI.Blank()
	terminalUI.Success("模块骨架已生成")
	terminalUI.Blank()
	terminalUI.Info("下一步：")
	terminalUI.Info("  1. 编辑 " + filePath + " 填写模板代码和 patch 逻辑")
	terminalUI.Info("  2. 运行 go build . 重新编译 Golider")
	terminalUI.Info("  3. 运行 Golider add " + moduleName + " 使用该模块")

	return nil
}

func validateModuleName(name string) error {
	if len(name) < 2 {
		return fmt.Errorf("模块名称至少 2 个字符")
	}
	for _, ch := range name {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '-' && ch != '_' {
			return fmt.Errorf("模块名称只能包含字母、数字、连字符和下划线")
		}
	}
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, "_") {
		return fmt.Errorf("模块名称不能以连字符或下划线开头")
	}
	return nil
}

func moduleNameToFileName(name string) string {
	result := strings.ReplaceAll(name, "-", "_")
	return result
}

func generateModuleSkeleton(moduleName string) string {
	identifier := moduleNameToFileName(moduleName)
	return strings.Join([]string{
		"package modules",
		"",
		"import (",
		"\t\"github.com/2859044775/Golider/internal/addon\"",
		")",
		"",
		"func init() {",
		"\taddon.RegisterModule(addon.ModuleDefinition{",
		"\t\tName: \"" + moduleName + "\",",
		"\t\tFiles: map[string]string{",
		"\t\t\t\"internal/http/" + identifier + ".go\": " + identifier + "Template,",
		"\t\t},",
		"\t\tBaseFiles: addon.CommonBaseFiles(),",
		"\t\tPatchFunc: func(targetDir string) error {",
		"\t\t\t// TODO: 在此处定义 patch 逻辑，例如：",
		"\t\t\t// addon.AppendEnvValue(targetDir, \"" + strings.ToUpper(strings.ReplaceAll(moduleName, "-", "_")) + "_ENABLED=true\")",
		"\t\t\t// addon.AddMiddlewareLine(targetDir, \"handler = " + identifier + "Middleware(handler)\")",
		"\t\t\t// addon.AddRouteLine(targetDir, \"mux.HandleFunc(\\\"/" + moduleName + "\\\", " + identifier + "Handler)\", \"/" + moduleName + "\")",
		"\t\t\treturn nil",
		"\t\t},",
		"\t})",
		"}",
		"",
		"const " + identifier + "Template = `package http",
		"",
		"// TODO: 实现 " + moduleName + " 模块代码",
		"`",
		"",
	}, "\n")
}
