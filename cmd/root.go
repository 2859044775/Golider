package cmd

import (
	"fmt"
	"strings"
)

func Execute(args []string) error {
	if len(args) == 0 {
		printRootUsage()
		return nil
	}

	switch args[0] {
	case "new":
		return runNew(args[1:])
	case "add":
		return runAdd(args[1:])
	case "verify":
		return runVerify(args[1:])
	case "verify-config":
		return runVerifyConfig(args[1:])
	case "doctor":
		return runDoctor(args[1:])
	case "version", "--version", "-v":
		return runVersion(args[1:])
	case "help", "-h", "--help":
		printRootUsage()
		return nil
	default:
		return fmt.Errorf("未知命令 %q，使用 `golider help` 查看可用命令", args[0])
	}
}

func printRootUsage() {
	lines := []string{
		"Golider：为 AI 时代生成生产可用的 Go 后端工程。",
		"",
		"用法：",
		"  golider <命令> [参数]",
		"",
		"可用命令：",
		"  new       生成一个最小可运行的 Go API 工程",
		"  add       为现有工程补充基础模块",
		"  verify    校验目标工程是否具备 Golider 最小结构",
		"  verify-config 校验目标工程的配置模板是否完整且有效",
		"  doctor    检查目标工程缺少哪些基础能力",
		"  version   输出当前 CLI 版本信息",
		"  help      查看帮助",
		"",
		"示例：",
		"  golider version",
		"  golider new demo --module github.com/acme/demo",
		"  golider add docker ./demo",
		"  golider add worker ./demo",
		"  golider add webhook ./demo",
		"  golider add auth ./demo",
		"  golider add postgres ./demo",
		"  golider add request-id ./demo",
		"  golider add timeout ./demo",
		"  golider add metrics ./demo",
		"  golider add rate-limit ./demo",
		"  golider add error-model ./demo",
		"  golider add cors ./demo",
		"  golider verify ./demo",
		"  golider verify-config ./demo",
		"  golider doctor ./demo",
		"  golider doctor fix ./demo",
	}

	fmt.Println(strings.Join(lines, "\n"))
}
