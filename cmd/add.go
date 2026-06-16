package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/2859044775/Golider/internal/addon"
)

func runAdd(args []string) error {
	parsed, err := parseAddArgs(args)
	if err != nil {
		return err
	}

	if parsed.ModuleName == "" {
		return fmt.Errorf("请提供要添加的模块，可用模块：%s", strings.Join(addon.List(), ", "))
	}

	targetDir := "."
	if parsed.TargetDir != "" {
		targetDir = filepath.Clean(parsed.TargetDir)
	}

	ui := newTerminalUI(os.Stdout)
	ui.Header("Golider 模块安装")
	ui.KeyValue("模块", parsed.ModuleName)
	ui.KeyValue("目标目录", targetDir)
	if parsed.Force {
		ui.KeyValue("覆盖模式", "是")
	}
	ui.Blank()

	ui.ProgressStep(1, 2, "正在安装模块")
	if err := addon.Install(addon.Options{
		ModuleName: parsed.ModuleName,
		TargetDir:  targetDir,
		Force:      parsed.Force,
	}); err != nil {
		return err
	}
	ui.Success("       " + parsed.ModuleName + " 已安装")
	ui.ProgressStep(2, 2, "安装完成")

	ui.Blank()
	ui.Success("模块 " + parsed.ModuleName + " 添加完成")
	return nil
}

type addArgs struct {
	ModuleName string
	TargetDir  string
	Force      bool
}

func parseAddArgs(args []string) (addArgs, error) {
	var result addArgs

	for _, arg := range args {
		switch arg {
		case "--force":
			result.Force = true
		default:
			if strings.HasPrefix(arg, "--") {
				return addArgs{}, fmt.Errorf("未知参数 %q", arg)
			}
			if result.ModuleName == "" {
				result.ModuleName = arg
				continue
			}
			if result.TargetDir == "" {
				result.TargetDir = arg
				continue
			}
			return addArgs{}, fmt.Errorf("参数过多，无法识别 %q", arg)
		}
	}

	return result, nil
}
