package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	// AuthToken is a bearer token to be provided in the Authorization header
	// to ensure requests are from a legitimate source.
	AuthToken string `yaml:"auth_token"`

	Mappings []ImageMapping `yaml:"mappings"`
}

// ImageMapping correlates a container image name to a Kubernetes deployment
// and namespace, so we know what to update for a given image.
type ImageMapping struct {
	ImageName      string   `yaml:"image"`
	DeploymentName string   `yaml:"deployment"`
	Namespace      string   `yaml:"namespace"`
	Providers      []string `yaml:"providers"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

// ProviderEnabled returns true if the given provider is listed in the config.
func ProviderEnabled(config *Config, provider string) bool {
	for _, p := range config.Mappings {
		for _, p2 := range p.Providers {
			if p2 == provider {
				return true
			}
		}
	}
	return false
}
