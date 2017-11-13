package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/sw/jstatbeat/beater"
)

func main() {
	err := beat.Run("jstatbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
