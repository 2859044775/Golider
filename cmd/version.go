package cmd

import (
	"fmt"
	"os"
)

var (
	Version   = "0.6.0"
	Commit    = "dev"
	BuildDate = "unknown"
)

func runVersion(_ []string) error {
	ui := newTerminalUI(os.Stdout)
	ui.Header("Golider 版本信息")
	fmt.Printf("Golider %s\n", Version)
	fmt.Printf("commit: %s\n", Commit)
	fmt.Printf("build_date: %s\n", BuildDate)
	return nil
}
