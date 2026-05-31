package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golider/golider/internal/check"
)

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	projectDir := "."
	if fs.NArg() > 0 {
		projectDir = filepath.Clean(fs.Arg(0))
	}

	ui := newTerminalUI(os.Stdout)
	ui.Header("Golider 结构校验")
	ui.KeyValue("目标目录", projectDir)

	missing := check.MissingItems(projectDir)
	if len(missing) > 0 {
		ui.Blank()
		ui.Failure("校验未通过，缺少以下文件：")
		for _, item := range missing {
			ui.StatusLine("缺失", item.Name+"："+item.Path)
		}
		return fmt.Errorf("目标工程不符合 Golider 最小结构")
	}

	ui.Blank()
	ui.Success("校验通过，目标工程具备 Golider 最小结构。")
	return nil
}
