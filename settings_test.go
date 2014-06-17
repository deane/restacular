package restacular

import (
	"testing"
	"time"
)

type tomlConfig struct {
	Title   string
	Owner   ownerInfo
	DB      database `toml:"database"`
	Servers map[string]server
	Clients clients
}

type ownerInfo struct {
	Name string
	Org  string `toml:"organization"`
	Bio  string
	DOB  time.Time
}

type database struct {
	Server  string
	Ports   []int
	ConnMax int `toml:"connection_max"`
	Enabled bool
}

type server struct {
	IP string
	DC string
}

type clients struct {
	Data  [][]interface{}
	Hosts []string
}

// Just test that loading a toml file works
func TestLoadingSettingsFromFile(t *testing.T) {
	var config tomlConfig
	err := LoadSettings("", "settings_test.toml", &config)

	if err != nil {
		t.Error("Errored when loading settings_test.toml file")
	}

	if config.Title != "TOML Example" {
		t.Error("Didn't load the settings file properly")
	}
}
