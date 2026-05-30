package cmd

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/golider/golider/internal/check"
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

	items := check.ConfigRequirements(projectDir)
	if len(items) == 0 {
		fmt.Println("未识别到需要校验的配置项，请先确认目标工程包含 .env.example。")
		return nil
	}

	fmt.Println("Golider 配置校验结果：")
	for _, item := range items {
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

	invalid := check.MissingOrInvalidConfig(projectDir)
	if len(invalid) == 0 {
		fmt.Println("结论：配置模板完整且值有效。")
		return nil
	}

	fmt.Printf("结论：发现 %d 项配置缺失或不合法。\n", len(invalid))
	return fmt.Errorf("目标工程配置模板校验未通过")
}
