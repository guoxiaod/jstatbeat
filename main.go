package main

import (
	"os"

	"github.com/guoxiaod/jstatbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
