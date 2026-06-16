package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/2859044775/Golider/internal/check"
)

func runVerifyConfig(args []string) error {
	fs := flag.NewFlagSet("verify-config", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	projectDir := "."
	if fs.NArg() > 0 {
		projectDir = filepath.Clean(fs.Arg(0))
	}

	ui := newTerminalUI(os.Stdout)
	ui.Header("Golider 配置校验")
	ui.KeyValue("目标目录", projectDir)

	items := check.ConfigRequirements(projectDir)
	if len(items) == 0 {
		ui.Blank()
		ui.Warning("未识别到需要校验的配置项，请先确认目标工程包含 .env.example。")
		return nil
	}

	ui.Blank()
	ui.Section("配置项")
	for _, item := range items {
		status := "正常"
		switch {
		case !item.Exists:
			status = "缺失"
		case !item.Valid:
			status = "无效"
		}

		if item.Value != "" {
			ui.StatusLine(status, item.Name+"："+item.Detail+"（当前值："+item.Value+"）")
			continue
		}
		ui.StatusLine(status, item.Name+"："+item.Detail)
	}

	invalid := check.MissingOrInvalidConfig(projectDir)
	if len(invalid) == 0 {
		ui.Blank()
		ui.Success("结论：配置模板完整且值有效。")
		return nil
	}

	ui.Blank()
	ui.Failure(fmt.Sprintf("结论：发现 %d 项配置缺失或不合法。", len(invalid)))
	return fmt.Errorf("目标工程配置模板校验未通过")
}
