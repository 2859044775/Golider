package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCreateModule(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "modules")

	err := runCreateModule([]string{"my-test-module", targetDir})
	if err != nil {
		t.Fatalf("create-module 失败: %v", err)
	}

	filePath := filepath.Join(targetDir, "my_test_module.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("读取生成的文件失败: %v", err)
	}

	source := string(content)

	// 检查关键结构
	checks := []struct {
		fragment string
		desc     string
	}{
		{"package modules", "package 声明"},
		{`"github.com/2859044775/Golider/internal/addon"`, "import addon 包"},
		{"func init()", "init 函数"},
		{"addon.RegisterModule(addon.ModuleDefinition{", "注册调用"},
		{`Name: "my-test-module"`, "模块名称"},
		{`"internal/http/my_test_module.go": my_test_moduleTemplate`, "文件映射"},
		{"addon.CommonBaseFiles()", "基础文件引用"},
		{"PatchFunc:", "patch 函数"},
		{"const my_test_moduleTemplate", "模板常量"},
	}

	for _, c := range checks {
		if !strings.Contains(source, c.fragment) {
			t.Fatalf("生成的文件缺少 %s，期望包含 %q", c.desc, c.fragment)
		}
	}
}

func TestCreateModuleAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "modules")

	// 第一次创建
	err := runCreateModule([]string{"existing-module", targetDir})
	if err != nil {
		t.Fatalf("第一次创建失败: %v", err)
	}

	// 第二次创建同名模块应该失败
	err = runCreateModule([]string{"existing-module", targetDir})
	if err == nil {
		t.Fatal("重复创建应该返回错误")
	}
	if !strings.Contains(err.Error(), "已存在") {
		t.Fatalf("错误信息应该包含'已存在'，实际: %v", err)
	}
}

func TestCreateModuleNoName(t *testing.T) {
	err := runCreateModule([]string{})
	if err == nil {
		t.Fatal("缺少模块名称应该返回错误")
	}
}

func TestCreateModuleInvalidName(t *testing.T) {
	cases := []string{
		"a",           // 太短
		"-invalid",    // 连字符开头
		"_invalid",    // 下划线开头
		"mod name",    // 包含空格
		"mod@name",    // 非法字符
		"",            // 空字符串
	}
	for _, name := range cases {
		err := runCreateModule([]string{name, t.TempDir()})
		if err == nil {
			t.Fatalf("模块名称 %q 应该被拒绝", name)
		}
	}
}

func TestModuleNameToFileName(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{"my-module", "my_module"},
		{"redis", "redis"},
		{"circuit-breaker", "circuit_breaker"},
		{"multi_tenant", "multi_tenant"},
	}
	for _, c := range cases {
		got := moduleNameToFileName(c.input)
		if got != c.expect {
			t.Fatalf("moduleNameToFileName(%q) = %q, 期望 %q", c.input, got, c.expect)
		}
	}
}
