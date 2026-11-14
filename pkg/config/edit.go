package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"go.yaml.in/yaml/v3"
)

// AddIgnoredApps adds names to ignoredApps in the config file, creating
// ~/.config/lightsout/config.yml if there is no config yet. Names already
// kept are skipped. It returns the names it added, in order.
//
// The file is rewritten, so YAML comments in it are lost. A config that can't
// be read is reported and left untouched.
func AddIgnoredApps(names []string) ([]string, error) {
	return editIgnoredApps(func(apps []string) ([]string, []string) {
		var added []string
		for _, name := range names {
			if !slices.Contains(apps, name) {
				apps = append(apps, name)
				added = append(added, name)
			}
		}
		return apps, added
	})
}

// RemoveIgnoredApps removes names from ignoredApps in the config file and
// returns the names it removed, in config order. Like AddIgnoredApps, it
// rewrites the file and loses its comments.
func RemoveIgnoredApps(names []string) ([]string, error) {
	return editIgnoredApps(func(apps []string) ([]string, []string) {
		var kept, removed []string
		for _, app := range apps {
			if slices.Contains(names, app) {
				removed = append(removed, app)
			} else {
				kept = append(kept, app)
			}
		}
		return kept, removed
	})
}

// editIgnoredApps loads the config (or an empty one), applies change to its
// ignoredApps, and writes the result back if anything changed. change
// returns the new list and the names it changed.
func editIgnoredApps(change func(apps []string) (updated, changed []string)) ([]string, error) {
	file, err := filePath()
	if err != nil {
		return nil, err
	}

	var conf Conf
	contents, err := os.ReadFile(file)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(contents, &conf); err != nil {
			return nil, fmt.Errorf("reading %s: %w", file, err)
		}
	case !os.IsNotExist(err):
		return nil, err
	}

	var changed []string
	conf.IgnoredApps, changed = change(conf.IgnoredApps)
	if len(changed) == 0 {
		return nil, nil
	}

	out, err := yaml.Marshal(conf)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return nil, err
	}
	return changed, os.WriteFile(file, out, 0o644)
}

// filePath returns the config file to edit: an existing config.yml or
// config.yaml, or config.yml if neither exists yet.
func filePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "lightsout")
	for _, name := range []string{"config.yml", "config.yaml"} {
		if info, err := os.Stat(filepath.Join(dir, name)); err == nil && !info.IsDir() {
			return filepath.Join(dir, name), nil
		}
	}
	return filepath.Join(dir, "config.yml"), nil
}
