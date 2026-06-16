package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/2859044775/Golider/internal/addon"
	"github.com/2859044775/Golider/internal/check"
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

	ui := newTerminalUI(os.Stdout)
	items := check.RequiredItems(projectDir)
	capabilities := check.Capabilities(projectDir)
	configItems := check.ConfigRequirements(projectDir)

	ui.Header("Golider 工程检查结果")
	ui.KeyValue("目标目录", projectDir)

	// 基础文件：正常折叠，异常展开
	normalFiles := 0
	abnormalFiles := 0
	for _, item := range items {
		if item.Exists {
			normalFiles++
		} else {
			abnormalFiles++
		}
	}
	ui.Blank()
	ui.Section("基础文件")
	ui.FoldedSummary(normalFiles, len(items), "项正常")
	for _, item := range items {
		if !item.Exists {
			ui.AbnormalItem("缺失", item.Name+" ("+item.Path+")")
		}
	}

	// 能力检查：正常折叠，异常展开
	normalCaps := 0
	abnormalCaps := 0
	for _, c := range capabilities {
		if c.Exists {
			normalCaps++
		} else {
			abnormalCaps++
		}
	}
	ui.Blank()
	ui.Section("能力检查")
	ui.FoldedSummary(normalCaps, len(capabilities), "项正常")
	for _, c := range capabilities {
		if !c.Exists {
			ui.AbnormalItem("缺失", c.Name+"："+c.Detail+" ("+c.Related+")")
		}
	}

	// 配置检查：正常折叠，异常展开
	normalCfg := 0
	abnormalCfg := 0
	if len(configItems) > 0 {
		for _, item := range configItems {
			if item.Exists && item.Valid {
				normalCfg++
			} else {
				abnormalCfg++
			}
		}
		ui.Blank()
		ui.Section("配置检查")
		ui.FoldedSummary(normalCfg, len(configItems), "项正常")
		for _, item := range configItems {
			if !item.Exists || !item.Valid {
				status := "缺失"
				if item.Exists && !item.Valid {
					status = "无效"
				}
				if item.Value != "" {
					ui.AbnormalItem(status, item.Name+"："+item.Detail+"（当前值："+item.Value+"）")
				} else {
					ui.AbnormalItem(status, item.Name+"："+item.Detail)
				}
			}
		}
	}

	totalItems := len(items) + len(capabilities) + len(configItems)
	totalNormal := normalFiles + normalCaps + normalCfg
	ui.ConclusionSummary(totalNormal, totalItems, abnormalFiles, abnormalCaps, abnormalCfg)

	if abnormalFiles > 0 || abnormalCaps > 0 || abnormalCfg > 0 {
		return fmt.Errorf("目标工程存在异常项，请检查或执行 `golider doctor fix`")
	}
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

	ui := newTerminalUI(os.Stdout)
	if err := ensureDoctorFixPrerequisites(projectDir); err != nil {
		return err
	}

	modules := modulesForDoctorFix(projectDir)
	if len(modules) == 0 {
		ui.Warning("当前工程没有可自动修复的通用能力。")
		return runDoctor([]string{projectDir})
	}

	ui.Header("Golider 自动修复")
	ui.KeyValue("目标目录", projectDir)
	ui.Blank()

	total := len(modules)
	for idx, moduleName := range modules {
		ui.ProgressStep(idx+1, total, "正在安装 "+moduleName)
		if err := addon.Install(addon.Options{
			ModuleName:   moduleName,
			TargetDir:    projectDir,
			SkipExisting: true,
		}); err != nil {
			ui.Failure("       " + moduleName + "：" + err.Error())
			continue
		}
		ui.Success("       " + moduleName)
	}

	ui.Blank()
	ui.Success("Golider 自动修复完成。")
	_ = runDoctor([]string{projectDir})
	return nil
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
