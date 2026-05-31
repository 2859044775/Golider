package cmd

import (
	"fmt"
	"os"
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
	ui := newTerminalUI(os.Stdout)
	ui.Header("Golider")
	ui.Info("为 AI 时代生成生产可用的 Go 后端工程。")
	ui.Blank()
	ui.Section("用法")
	ui.KeyValue("命令", "golider <命令> [参数]")
	ui.Blank()
	ui.Section("可用命令")
	ui.KeyValue("new", "生成一个最小可运行的 Go API 工程")
	ui.KeyValue("add", "为现有工程补充基础模块")
	ui.KeyValue("verify", "校验目标工程是否具备 Golider 最小结构")
	ui.KeyValue("verify-config", "校验目标工程的配置模板是否完整且有效")
	ui.KeyValue("doctor", "检查目标工程缺少哪些基础能力")
	ui.KeyValue("version", "输出当前 CLI 版本信息")
	ui.KeyValue("help", "查看帮助")
	ui.Blank()
	ui.Section("示例")
	ui.KeyValue("示例", "golider version")
	ui.KeyValue("示例", "golider new demo --module github.com/acme/demo")
	ui.KeyValue("示例", "golider add docker ./demo")
	ui.KeyValue("示例", "golider add worker ./demo")
	ui.KeyValue("示例", "golider add webhook ./demo")
	ui.KeyValue("示例", "golider add auth ./demo")
	ui.KeyValue("示例", "golider add postgres ./demo")
	ui.KeyValue("示例", "golider add request-id ./demo")
	ui.KeyValue("示例", "golider add timeout ./demo")
	ui.KeyValue("示例", "golider add metrics ./demo")
	ui.KeyValue("示例", "golider add rate-limit ./demo")
	ui.KeyValue("示例", "golider add error-model ./demo")
	ui.KeyValue("示例", "golider add cors ./demo")
	ui.KeyValue("示例", "golider verify ./demo")
	ui.KeyValue("示例", "golider verify-config ./demo")
	ui.KeyValue("示例", "golider doctor ./demo")
	ui.KeyValue("示例", "golider doctor fix ./demo")
}
