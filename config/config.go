// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period time.Duration `config:"period"`
	Name   string        `config:"name"`
	Freq   string        `config:"freq"`
}

var DefaultConfig = Config{
	Period: 1 * time.Second,
	Name:   "JConsole",
	Freq:   "1000",
}
