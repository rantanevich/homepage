package watcher

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rantanevich/homepage/app/config/dynamic"
)

func mergeConfig(configs map[string]dynamic.Config) dynamic.Config {
	conf := make(dynamic.Config)
	for _, groups := range configs {
		for groupName, services := range groups {
			for serviceName, service := range services {
				if _, exists := conf[groupName]; !exists {
					conf[groupName] = make(map[string]dynamic.Service)
				}
				service.Icon = ResolveIcon(service.Icon)
				conf[groupName][serviceName] = service
			}
		}
	}
	return conf
}

func ResolveIcon(path string) string {
	if strings.HasPrefix(path, "http") || strings.HasPrefix(path, "/") {
		return path
	}

	ext := filepath.Ext(path)
	if ext == "" {
		ext = ".png"
	}
	format := ext[1:]
	name := strings.TrimSuffix(path, ext)

	if strings.HasPrefix(name, "sh-") {
		name = strings.TrimPrefix(name, "sh-")
		return fmt.Sprintf("https://cdn.jsdelivr.net/gh/selfhst/icons/%s/%s.%s", format, name, format)
	}

	return fmt.Sprintf("https://cdn.jsdelivr.net/gh/homarr-labs/dashboard-icons/%s/%s.%s", format, name, format)
}
