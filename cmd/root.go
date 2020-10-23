package cmd

import (
	"github.com/guoxiaod/jstatbeat/beater"

	cmd "github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
)

var name = "jstatbeat"
var version = "1.0.0" // TODO consider moving this or pulling from conf or env

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmdWithSettings(beater.New, instance.Settings{Name: name, Version: version})
