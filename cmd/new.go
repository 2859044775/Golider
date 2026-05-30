package cmd

import (
	"fmt"
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

	fmt.Printf("已生成项目 %s\n", opts.AppName)
	fmt.Printf("下一步：\n")
	fmt.Printf("  cd %s\n", targetDir)
	fmt.Printf("  go run ./cmd/api\n")
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
