package dynamic

type Config map[string]map[string]Service

type Service struct {
	URL         string `yaml:"url"`
	Icon        string `yaml:"icon"`
	Description string `yaml:"description"`
}

type Message struct {
	ProviderName string
	Config       Config
}

func (c Config) DeepCopy() Config {
	clone := make(Config)
	for groupName, services := range c {
		clone[groupName] = make(map[string]Service)
		for serviceName, service := range services {
			clone[groupName][serviceName] = service
		}
	}
	return clone
}
