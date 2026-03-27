package main

import (
	"fmt"
	"os"

	"github.com/crper/tqrx/internal/cli"
)

// main 保持进程级装配最小化，并把所有行为委托给 CLI runner。
func main() {
	runner := cli.NewRunner()
	if err := runner.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
