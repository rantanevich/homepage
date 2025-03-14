package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port       int        `yaml:"port"`
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

func Load() (*Config, error) {
	confPath := os.Getenv("CONFIG_PATH")
	if confPath == "" {
		confPath = "config.yml"
	}

	file, err := os.Open(confPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s file: %w", confPath, err)
	}
	defer file.Close()

	conf := &Config{}
	if err := yaml.NewDecoder(file).Decode(conf); err != nil {
		return nil, fmt.Errorf("failed to decode %s file: %w", confPath, err)
	}

	if err := conf.setDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	return conf, nil
}

func (c *Config) setDefaults() error {
	if c.Port == 0 {
		c.Port = 3000
	}

	if c.Title == "" {
		c.Title = "Homepage"
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

	if c.Logo == "" {
		c.Logo = findIcon(userIcons, staticIcons, "logo.png")
	} else {
		c.Logo = findIcon(userIcons, staticIcons, c.Logo)
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

			c.Categories[i].Services[j].Icon = findIcon(userIcons, staticIcons, service.Icon)
		}
	}
	return nil
}
