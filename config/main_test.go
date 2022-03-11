package config

import (
	"testing"
)

func TestLoadConfigValid(t *testing.T) {
	config, err := LoadConfig("fixtures/config.valid.yaml")
	if err != nil {
		t.Errorf("Error while loading config fixture: %v", err)
	}

	if config.AuthToken != "abc123" {
		t.Errorf("LoadConfig parsed AuthToken incorrectly. Got: %v", config.AuthToken)
	}
	if config.Mappings[0].ImageName != "watashi/app" {
		t.Errorf("LoadConfig parsed Mapping.ImageName incorrectly. Got: %v", config.Mappings[0].ImageName)
	}
	if config.Mappings[0].DeploymentName != "abc" {
		t.Errorf("LoadConfig parsed Mapping.DeploymentName incorrectly. Got: %v", config.Mappings[0].DeploymentName)
	}
	if config.Mappings[0].Namespace != "default" {
		t.Errorf("LoadConfig parsed Mapping.Namespace incorrectly. Got: %v", config.Mappings[0].Namespace)
	}
}
