package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golider/golider/internal/addon"
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

	if err := addon.Install(addon.Options{
		ModuleName: parsed.ModuleName,
		TargetDir:  targetDir,
		Force:      parsed.Force,
	}); err != nil {
		return err
	}

	ui := newTerminalUI(os.Stdout)
	ui.Header("模块添加完成")
	ui.Success("已添加模块 " + parsed.ModuleName)
	ui.KeyValue("目标目录", targetDir)
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
