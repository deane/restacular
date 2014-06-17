package restacular

import (
	"github.com/BurntSushi/toml"
)

// Just a wrapper around the toml decoder for now
func LoadSettings(envPrefix string, path string, settings interface{}) error {
	_, err := toml.DecodeFile(path, settings)
	return err
}
