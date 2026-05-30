package cmd

import "fmt"

var (
	Version   = "0.1.0"
	Commit    = "dev"
	BuildDate = "unknown"
)

func runVersion(_ []string) error {
	fmt.Printf("golider %s\n", Version)
	fmt.Printf("commit: %s\n", Commit)
	fmt.Printf("build_date: %s\n", BuildDate)
	return nil
}
