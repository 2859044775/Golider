package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golider/golider/internal/scaffold"
)

func runNew(args []string) error {
	opts, err := parseNewArgs(args)
	if err != nil {
		return err
	}

	if strings.TrimSpace(opts.AppName) == "" {
		return fmt.Errorf("请提供项目名，例如 `golider new demo --module github.com/acme/demo`")
	}

	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = filepath.Clean(opts.AppName)
	}

	ui := newTerminalUI(os.Stdout)
	ui.Header("Golider 项目生成")
	ui.KeyValue("项目名", opts.AppName)
	ui.KeyValue("模块路径", opts.Module)
	ui.KeyValue("目标目录", targetDir)
	ui.KeyValue("默认端口", opts.Port)
	ui.Blank()

	ui.ProgressStep(1, 4, "初始化目录结构")

	scaffoldOpts := scaffold.Options{
		AppName:     opts.AppName,
		Module:      opts.Module,
		TargetDir:   targetDir,
		DefaultPort: opts.Port,
		Force:       opts.Force,
	}

	if err := scaffold.CreateProject(scaffoldOpts); err != nil {
		return err
	}

	ui.Success("       " + "目录结构已就绪")
	ui.ProgressStep(2, 4, "写入工程文件")
	ui.Success("       " + "工程文件已生成（go.mod、Dockerfile、Makefile 等）")
	ui.ProgressStep(3, 4, "注入默认能力")
	ui.Success("       " + "默认能力已注入（日志、中间件、路由、仓储、服务层）")
	ui.ProgressStep(4, 4, "生成完成")

	ui.Blank()
	ui.Header("项目生成完成")
	ui.Success("已生成项目 " + opts.AppName)
	ui.Blank()
	ui.Section("下一步")
	ui.KeyValue("进入目录", "cd "+targetDir)
	ui.KeyValue("复制环境变量", "cp .env.example .env")
	ui.KeyValue("启动服务", "go run ./cmd/api")

	return nil
}

type newArgs struct {
	AppName   string
	Module    string
	TargetDir string
	Port      string
	Force     bool
}

func parseNewArgs(args []string) (newArgs, error) {
	result := newArgs{
		Port: "8080",
	}

	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		switch arg {
		case "--module":
			value, next, err := requireNextValue(args, idx, arg)
			if err != nil {
				return newArgs{}, err
			}
			result.Module = value
			idx = next
		case "--target":
			value, next, err := requireNextValue(args, idx, arg)
			if err != nil {
				return newArgs{}, err
			}
			result.TargetDir = value
			idx = next
		case "--port":
			value, next, err := requireNextValue(args, idx, arg)
			if err != nil {
				return newArgs{}, err
			}
			result.Port = value
			idx = next
		case "--force":
			result.Force = true
		default:
			if strings.HasPrefix(arg, "--") {
				return newArgs{}, fmt.Errorf("未知参数 %q", arg)
			}
			if result.AppName != "" {
				return newArgs{}, fmt.Errorf("只允许提供一个项目名，收到多余参数 %q", arg)
			}
			result.AppName = arg
		}
	}

	return result, nil
}

func requireNextValue(args []string, current int, flagName string) (string, int, error) {
	next := current + 1
	if next >= len(args) {
		return "", current, fmt.Errorf("参数 %s 缺少值", flagName)
	}
	if strings.HasPrefix(args[next], "--") {
		return "", current, fmt.Errorf("参数 %s 缺少值", flagName)
	}
	return args[next], next, nil
}
