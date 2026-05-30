package cmd

import (
	"flag"
	"fmt"
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

	missing := check.MissingItems(projectDir)
	if len(missing) > 0 {
		fmt.Println("校验未通过，缺少以下文件：")
		for _, item := range missing {
			fmt.Printf("- %s：%s\n", item.Name, item.Path)
		}
		return fmt.Errorf("目标工程不符合 Golider 最小结构")
	}

	fmt.Println("校验通过，目标工程具备 Golider 最小结构。")
	return nil
}
