package config

import (
	"os"

	"github.com/gleich/statuser/v2"
	"go.yaml.in/yaml/v3"
)

// Directory of the config file, relative to the user's home directory
const configDir = "/.config/lightsout/"

// Full config file path, set by Exists and read by Read
var path string

// Config for lightsout
type Conf struct {
	IgnoreUpdates bool     `yaml:"ignoredUpdates"`
	IgnoredApps   []string `yaml:"ignoredApps"`
}

// Check if the config exists
func Exists() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		statuser.Error("Failed to get user home path", err, 1)
	}
	dir := home + configDir

	for _, name := range []string{"config.yml", "config.yaml"} {
		candidate := dir + name
		info, err := os.Stat(candidate)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			statuser.Error("Failed to check for config in\n\t"+candidate, err, 1)
		}
		if !info.IsDir() {
			path = candidate
			return true
		}
	}
	return false
}

// Read from the config file
func Read(conf *Conf) *Conf {
	contents, err := os.ReadFile(path)
	if err != nil {
		statuser.Error("Failed to read from config in\n\t"+path, err, 1)
	}

	err = yaml.Unmarshal(contents, &conf)
	if err != nil {
		statuser.Error("Failed to read the config", err, 1)
	}
	return conf
}
