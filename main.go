package main

import (
	"fmt"
	"os"

	"github.com/2859044775/Golider/cmd"

	// 注册内置 addon 模块
	_ "github.com/2859044775/Golider/internal/addon/modules"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
