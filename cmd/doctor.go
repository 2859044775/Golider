package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golider/golider/internal/addon"
	"github.com/golider/golider/internal/check"
)

func runDoctor(args []string) error {
	if len(args) > 0 && args[0] == "fix" {
		return runDoctorFix(args[1:])
	}

	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	projectDir := "."
	if fs.NArg() > 0 {
		projectDir = filepath.Clean(fs.Arg(0))
	}

	items := check.RequiredItems(projectDir)
	fmt.Println("Golider 工程检查结果：")
	fmt.Println("基础文件：")
	for _, item := range items {
		status := "缺失"
		if item.Exists {
			status = "正常"
		}
		fmt.Printf("- [%s] %s (%s)\n", status, item.Name, item.Path)
	}

	fmt.Println("能力检查：")
	capabilities := check.Capabilities(projectDir)
	for _, capability := range capabilities {
		status := "缺失"
		if capability.Exists {
			status = "正常"
		}
		fmt.Printf("- [%s] %s：%s (%s)\n", status, capability.Name, capability.Detail, capability.Related)
	}

	configItems := check.ConfigRequirements(projectDir)
	if len(configItems) > 0 {
		fmt.Println("配置检查：")
		for _, item := range configItems {
			status := "正常"
			switch {
			case !item.Exists:
				status = "缺失"
			case !item.Valid:
				status = "无效"
			}

			if item.Value != "" {
				fmt.Printf("- [%s] %s：%s（当前值：%s）\n", status, item.Name, item.Detail, item.Value)
				continue
			}
			fmt.Printf("- [%s] %s：%s\n", status, item.Name, item.Detail)
		}
	}

	missing := check.MissingItems(projectDir)
	missingCapabilities := check.MissingCapabilities(projectDir)
	invalidConfig := check.MissingOrInvalidConfig(projectDir)
	if len(missing) == 0 && len(missingCapabilities) == 0 && len(invalidConfig) == 0 {
		fmt.Println("结论：当前工程已经具备首版最小能力。")
		return nil
	}

	fmt.Printf("结论：当前工程还缺少 %d 项基础文件、%d 项能力，另有 %d 项配置缺失或无效。\n", len(missing), len(missingCapabilities), len(invalidConfig))
	return nil
}

func runDoctorFix(args []string) error {
	fs := flag.NewFlagSet("doctor fix", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	projectDir := "."
	if fs.NArg() > 0 {
		projectDir = filepath.Clean(fs.Arg(0))
	}

	if err := ensureDoctorFixPrerequisites(projectDir); err != nil {
		return err
	}

	modules := modulesForDoctorFix(projectDir)
	if len(modules) == 0 {
		fmt.Println("当前工程没有可自动修复的通用能力。")
		return runDoctor([]string{projectDir})
	}

	fmt.Println("Golider 自动修复开始：")
	for _, moduleName := range modules {
		if err := addon.Install(addon.Options{
			ModuleName:   moduleName,
			TargetDir:    projectDir,
			SkipExisting: true,
		}); err != nil {
			fmt.Printf("- [失败] %s：%v\n", moduleName, err)
			continue
		}
		fmt.Printf("- [完成] %s\n", moduleName)
	}

	fmt.Println("Golider 自动修复完成。")
	return runDoctor([]string{projectDir})
}

func modulesForDoctorFix(projectDir string) []string {
	var modules []string

	for _, item := range check.MissingItems(projectDir) {
		switch item.Path {
		case ".gitignore":
			modules = appendIfMissing(modules, "gitignore")
		case "Dockerfile":
			modules = appendIfMissing(modules, "docker")
		case ".github/workflows/ci.yml":
			modules = appendIfMissing(modules, "ci")
		}
	}

	for _, capability := range check.MissingCapabilities(projectDir) {
		switch capability.Name {
		case "统一错误模型":
			modules = appendIfMissing(modules, "error-model")
		case "请求标识":
			modules = appendIfMissing(modules, "request-id")
		case "请求超时":
			modules = appendIfMissing(modules, "timeout")
		case "指标采集":
			modules = appendIfMissing(modules, "metrics")
		case "限流保护":
			modules = appendIfMissing(modules, "rate-limit")
		case "跨域支持":
			modules = appendIfMissing(modules, "cors")
		}
	}

	return modules
}

func appendIfMissing(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func ensureDoctorFixPrerequisites(projectDir string) error {
	envPath := filepath.Join(projectDir, ".env.example")
	if _, err := os.Stat(envPath); err == nil {
		return nil
	}

	content := "PORT=8080\nSHUTDOWN_TIMEOUT=10s\nLOG_LEVEL=info\n"
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("创建 .env.example 失败：%w", err)
	}
	return nil
}
