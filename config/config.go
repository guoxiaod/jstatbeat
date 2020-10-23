// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period       time.Duration `config:"period"`
	Freq         string        `config:"freq"`
	Names        []string      `config:"names"`
	ExcludeNames []string      `config:"exclude_names"`
}

var DefaultConfig = Config{
	Period:       1 * time.Second,
	Freq:         "1000",
	Names:        []string{},
	ExcludeNames: []string{},
}
