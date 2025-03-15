package static

import (
	"log"
	"os"

	"github.com/traefik/paerser/env"

	"github.com/rantanevich/homepage/app/provider/docker"
	"github.com/rantanevich/homepage/app/provider/file"
)

type Config struct {
	Port      int
	Title     string
	Logo      string
	IconsDir  string
	Providers Providers
}

type Providers struct {
	File   *file.Provider
	Docker *docker.Provider `label:"allowEmpty"`
}

func (c *Config) SetDefaults() {
	if c.Port == 0 {
		c.Port = 3000
	}

	if c.Title == "" {
		c.Title = "Homepage"
	}

	if c.Logo == "" {
		c.Logo = "/static/icons/logo.png"
	}

	if c.IconsDir == "" {
		c.IconsDir = "/icons"
	}
}

func Load() *Config {
	conf := &Config{}
	if err := env.Decode(os.Environ(), "HOMEPAGE_", conf); err != nil {
		log.Fatalf("[FATAL] cannot load config: %v", err)
	}
	conf.SetDefaults()
	return conf
}
