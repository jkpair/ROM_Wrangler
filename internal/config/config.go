package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SourceDirs    []string          `yaml:"source_dirs"`
	ChdmanPath    string            `yaml:"chdman_path,omitempty"`
	DeleteArchive bool              `yaml:"delete_archive,omitempty"`
	Device        DeviceConfig      `yaml:"device"`
	Scraping      ScrapingConfig    `yaml:"scraping"`
	Transfer      TransferConfig    `yaml:"transfer"`
	Aliases       map[string]string `yaml:"aliases,omitempty"`
}

type DeviceConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	RootPath string `yaml:"root_path"`
}

type ScrapingConfig struct {
	ScreenScraperUser string   `yaml:"screenscraper_user,omitempty"`
	ScreenScraperPass string   `yaml:"screenscraper_pass,omitempty"`
	DATDirs           []string `yaml:"dat_dirs,omitempty"`
}

type TransferConfig struct {
	Method      string `yaml:"method"`
	SyncMode    bool   `yaml:"sync_mode"`
	USBPath     string `yaml:"usb_path,omitempty"`
	Concurrency int    `yaml:"concurrency"`
}

func DefaultConfig() *Config {
	return &Config{
		Device: DeviceConfig{
			Type:     "replayos",
			Host:     "replayos.local",
			Port:     22,
			User:     "root",
			Password: "replayos",
			RootPath: "/",
		},
		Transfer: TransferConfig{
			Method:      "sftp",
			SyncMode:    true,
			Concurrency: 1,
		},
	}
}

func DefaultPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "romwrangler", "config.yaml")
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := DefaultConfig()
			if saveErr := Save(cfg, path); saveErr != nil {
				return cfg, nil // return defaults even if save fails
			}
			return cfg, nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.expandPaths()
	return cfg, nil
}

// expandPaths resolves ~ to the user's home directory in all path fields.
func (cfg *Config) expandPaths() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	for i, d := range cfg.SourceDirs {
		cfg.SourceDirs[i] = expandTilde(d, home)
	}
	cfg.ChdmanPath = expandTilde(cfg.ChdmanPath, home)
	cfg.Device.RootPath = expandTilde(cfg.Device.RootPath, home)
	cfg.Transfer.USBPath = expandTilde(cfg.Transfer.USBPath, home)
	for i, d := range cfg.Scraping.DATDirs {
		cfg.Scraping.DATDirs[i] = expandTilde(d, home)
	}
}

func expandTilde(path, home string) string {
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

// ROMDirs returns the ROM directories (SourceDirs[i]/roms) for each root.
func (cfg *Config) ROMDirs() []string {
	dirs := make([]string, len(cfg.SourceDirs))
	for i, d := range cfg.SourceDirs {
		dirs[i] = filepath.Join(d, "roms")
	}
	return dirs
}

func Save(cfg *Config, path string) error {
	if path == "" {
		path = DefaultPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
