package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ExcludePatterns  []string `toml:"exclude_patterns"`
	MaxFileSize      int64    `toml:"max_file_size"`
	RespectGitignore bool     `toml:"respect_gitignore"`
	Format           string   `toml:"format"`
}

var defaultExcludes = []string{
	".git", "node_modules", ".venv", "venv", "__pycache__",
	".DS_Store", "dist", "build", "target",
}

func Default() Config {
	return Config{
		ExcludePatterns:  append([]string(nil), defaultExcludes...),
		MaxFileSize:      1 << 20, // 1 MiB
		RespectGitignore: true,
		Format:           "tail",
	}
}

// Load reads the config from ~/.config/ccopy/config.toml, falling back to defaults
// if the file does not exist. Explicit fields in the file override the defaults.
func Load() (Config, error) {
	path, err := defaultPath()
	if err != nil {
		return Default(), err
	}
	return LoadFrom(path)
}

// LoadFrom reads the config from the given path. Returns defaults (and no error)
// if the file does not exist.
func LoadFrom(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	var file Config
	// Decode into a zeroed value so we can tell which fields were set.
	md, err := toml.Decode(string(data), &file)
	if err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	for _, key := range md.Keys() {
		switch key.String() {
		case "exclude_patterns":
			cfg.ExcludePatterns = file.ExcludePatterns
		case "max_file_size":
			cfg.MaxFileSize = file.MaxFileSize
		case "respect_gitignore":
			cfg.RespectGitignore = file.RespectGitignore
		case "format":
			cfg.Format = file.Format
		}
	}
	return cfg, nil
}

func defaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "ccopy", "config.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ccopy", "config.toml"), nil
}

// ParseSize accepts values like "5MB", "512KB", "1048576". Case-insensitive.
// Uses binary (1024) units for KB/MB/GB to match common CLI conventions.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, errors.New("empty size")
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "GB"):
		mult = 1 << 30
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		mult = 1 << 20
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		mult = 1 << 10
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}
	s = strings.TrimSpace(s)
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size: %w", err)
	}
	if n < 0 {
		return 0, errors.New("size must be non-negative")
	}
	return int64(n * float64(mult)), nil
}
