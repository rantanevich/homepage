package docker

import (
	"github.com/docker/docker/api/types/container"

	"github.com/rantanevich/homepage/app/config/dynamic"
)

var requiredLabels = []string{
	"homepage.group",
	"homepage.service",
	"homepage.url",
}

func buildConfig(containers []container.Summary) dynamic.Config {
	conf := make(dynamic.Config)

	for _, container := range containers {
		if !keepContainer(container) {
			continue
		}

		group := container.Labels["homepage.group"]
		service := container.Labels["homepage.service"]

		if _, exists := conf[group]; !exists {
			conf[group] = make(map[string]dynamic.Service)
		}

		conf[group][service] = dynamic.Service{
			URL:         container.Labels["homepage.url"],
			Icon:        container.Labels["homepage.icon"],
			Description: container.Labels["homepage.description"],
		}
	}

	return conf
}

func keepContainer(container container.Summary) bool {
	if container.Labels == nil {
		return false
	}

	for _, name := range requiredLabels {
		value, exists := container.Labels[name]
		if !exists || value == "" {
			return false
		}
	}

	return true
}
