package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// knownIWADs lists the standard IWAD filenames that the scanner looks for.
var knownIWADs = []string{
	"doom.wad",
	"doom2.wad",
	"plutonia.wad",
	"tnt.wad",
	"heretic.wad",
	"hexen.wad",
	"strife1.wad",
	"freedoom1.wad",
	"freedoom2.wad",
}

// Config holds persistent settings for the application.
type Config struct {
	DSDADoomPath     string   `json:"dsda_doom_path"`     // Path to dsda-doom.exe
	IWADPaths        []string `json:"iwad_paths"`         // Full paths to registered IWAD files
	IWADFolder       string   `json:"iwad_folder"`        // Folder to auto-scan for IWADs
	DefaultDemosDir  string   `json:"default_demos_dir"`  // Default folder for .lmp browsing
	DefaultOutputDir string   `json:"default_output_dir"` // Default folder for generated .bat files
}

// ScanIWADFolder scans the configured IWAD folder for known IWAD files
// and updates IWADPaths with the results. Returns the number of IWADs found.
func (c *Config) ScanIWADFolder() (int, error) {
	if c.IWADFolder == "" {
		return 0, nil
	}

	entries, err := os.ReadDir(c.IWADFolder)
	if err != nil {
		return 0, err
	}

	found := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		for _, known := range knownIWADs {
			if name == known {
				found = append(found, filepath.Join(c.IWADFolder, e.Name()))
				break
			}
		}
	}

	c.IWADPaths = found
	return len(found), nil
}

// configPath returns the path to the config file next to the executable.
func configPath() string {
	exe, err := os.Executable()
	if err != nil {
		// Fallback: use current working directory
		if wd, wdErr := os.Getwd(); wdErr == nil {
			return filepath.Join(wd, "config.json")
		}
		return "config.json"
	}
	// Resolve symlinks so we get the real location
	real, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return filepath.Join(filepath.Dir(exe), "config.json")
	}
	return filepath.Join(filepath.Dir(real), "config.json")
}

// LoadConfig reads the config from disk. Returns a default config if the file doesn't exist.
func LoadConfig() *Config {
	cfg := &Config{}

	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist yet — that's fine, return defaults
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		// JSON is corrupt — return defaults rather than crashing
		return &Config{}
	}
	return cfg
}

// SaveConfig writes the config to disk. Returns the path it was saved to.
func SaveConfig(cfg *Config) (string, error) {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	path := configPath()
	return path, os.WriteFile(path, data, 0644)
}

// IWADNames returns just the base filenames from the configured IWAD paths.
func (c *Config) IWADNames() []string {
	names := make([]string, len(c.IWADPaths))
	for i, p := range c.IWADPaths {
		names[i] = filepath.Base(p)
	}
	return names
}

// IWADOptions returns IWAD names for dropdowns. Falls back to defaults if none configured.
func (c *Config) IWADOptions() []string {
	if len(c.IWADPaths) > 0 {
		return c.IWADNames()
	}
	return []string{"doom.wad", "doom2.wad", "plutonia.wad", "tnt.wad"}
}
