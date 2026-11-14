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

// Conf is the contents of the config file.
type Conf struct {
	IgnoreUpdates bool     `yaml:"ignoredUpdates,omitempty"`
	IgnoredApps   []string `yaml:"ignoredApps"`
}

// Exists reports whether a config file exists, and remembers its path for
// Read. It exits lightsout if the home folder or the file can't be checked.
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

// Read reads the config file found by Exists into conf. It exits lightsout
// if the file can't be read or isn't valid YAML.
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
