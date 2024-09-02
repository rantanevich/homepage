package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Title      string     `yaml:"title"`
	Logo       string     `yaml:"logo"`
	IconsDir   string     `yaml:"icons"`
	Categories []Category `yaml:"categories"`
}

type Category struct {
	Name     string    `yaml:"name"`
	Services []Service `yaml:"services"`
}

type Service struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
	Icon        string `yaml:"icon"`
}

const (
	staticIconsDir     = "./web/static/icons"
	staticPattern      = "/static/"
	iconsPattern       = "/icons/"
	staticIconsPattern = staticPattern + "/icons/"
	defaultIcon        = staticIconsPattern + "no-icon.svg"
)

func LoadConfig() (*Config, error) {
	confPath := os.Getenv("CONFIG_PATH")
	if confPath == "" {
		confPath = "config.yml"
	}

	f, err := os.Open(confPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open config file (%s): %w", confPath, err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("cannot decode config file (%s): %w", confPath, err)
	}

	if err := cfg.setDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set default values: %w", err)
	}

	return &cfg, nil
}

func (c *Config) setDefaults() error {
	if c.Title == "" {
		c.Title = "Homepage"
	}

	if c.Logo == "" {
		c.Logo = filepath.Join(staticIconsPattern, "logo.png")
	} else {
		c.Logo = filepath.Join(iconsPattern, c.Logo)
	}

	if c.IconsDir == "" {
		c.IconsDir = "/icons"
	}

	staticIcons, err := getStaticIcons()
	if err != nil {
		return err
	}

	userIcons, err := getUserIcons(c.IconsDir)
	if err != nil {
		return err
	}

	for i, category := range c.Categories {
		if category.Name == "" {
			return fmt.Errorf("categories[%d].name field is required", i)
		}

		for j, service := range category.Services {
			if service.Name == "" {
				return fmt.Errorf("categories[%d].services[%d].name field is required", i, j)
			}

			if service.URL == "" {
				return fmt.Errorf("categories[%d].services[%d].url field is required", i, j)
			}

			c.Categories[i].Services[j].Icon = c.findIcon(userIcons, staticIcons, service.Icon)
		}
	}
	return nil
}

func (c *Config) findIcon(userIcons, staticIcons []string, path string) string {
	if path == "" {
		return defaultIcon
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	ext := filepath.Ext(path)
	if ext != "" {
		if slices.Contains(userIcons, path) {
			return filepath.Join(iconsPattern, path)
		}

		if slices.Contains(staticIcons, path) {
			return filepath.Join(staticIconsPattern, path)
		}
		return defaultIcon
	}

	for _, ext := range []string{".svg", ".png"} {
		fname := path + ext
		if slices.Contains(userIcons, fname) {
			return filepath.Join(iconsPattern, fname)
		}

		if slices.Contains(staticIcons, fname) {
			return filepath.Join(staticIconsPattern, fname)
		}
	}
	return defaultIcon
}
