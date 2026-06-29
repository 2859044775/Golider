package addon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegisterAndInstallCustomModule(t *testing.T) {
	// 保存原始状态
	originalRegistered := registeredModules
	registeredModules = nil
	defer func() { registeredModules = originalRegistered }()

	// 注册一个自定义模块
	RegisterModule(ModuleDefinition{
		Name: "test-custom",
		Files: map[string]string{
			"internal/http/test_custom.go": "package http\n\n// test custom module\n",
		},
		BaseFiles: map[string]string{
			"internal/observability/logger.go": "package observability\n\n// mock logger\n",
		},
		PatchFunc: func(targetDir string) error {
			return AppendEnvValue(targetDir, "TEST_CUSTOM_ENABLED=true")
		},
	})

	// 验证 availableModules 包含自定义模块
	modules := availableModules()
	found := false
	for _, m := range modules {
		if m == "test-custom" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("availableModules 未包含自定义模块 test-custom")
	}

	// 创建目标项目
	projectDir := t.TempDir()
	writeTestFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeTestFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")

	// 安装自定义模块
	err := Install(Options{
		ModuleName: "test-custom",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装自定义模块失败: %v", err)
	}

	// 验证文件已生成
	moduleFile := filepath.Join(projectDir, "internal", "http", "test_custom.go")
	if _, err := os.Stat(moduleFile); os.IsNotExist(err) {
		t.Fatal("自定义模块文件未生成")
	}

	loggerFile := filepath.Join(projectDir, "internal", "observability", "logger.go")
	if _, err := os.Stat(loggerFile); os.IsNotExist(err) {
		t.Fatal("基础文件 logger.go 未生成")
	}

	// 验证 patch 已执行
	envContent := readTestFile(t, filepath.Join(projectDir, ".env.example"))
	if !strings.Contains(envContent, "TEST_CUSTOM_ENABLED=true") {
		t.Fatalf(".env.example 未追加 TEST_CUSTOM_ENABLED: %s", envContent)
	}
}

func TestRegisterModuleNotOverrideBuiltin(t *testing.T) {
	// 保存原始状态
	originalRegistered := registeredModules
	registeredModules = nil
	defer func() { registeredModules = originalRegistered }()

	// 注册一个与内置模块同名的模块
	RegisterModule(ModuleDefinition{
		Name: "redis",
		Files: map[string]string{
			"internal/http/evil.go": "package http\n\n// evil override\n",
		},
		BaseFiles: map[string]string{},
		PatchFunc: nil,
	})

	// moduleFiles 应该返回内置模块的文件，而不是自定义的
	files := moduleFiles("redis")
	if _, ok := files["internal/http/evil.go"]; ok {
		t.Fatal("自定义模块不应该覆盖内置模块的文件")
	}
	if _, ok := files["internal/store/redis.go"]; !ok {
		t.Fatal("内置 redis 模块文件应该存在")
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	return string(content)
}
