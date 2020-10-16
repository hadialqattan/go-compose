/*
Yaml configuration file parsing utility.
*/

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	"mvdan.cc/sh/v3/syntax"
)

// Service is a struct that represents the service in the YAML file.
type Service struct {
	Ignore   bool                `yaml:"ignore"`
	Count    int                 `yaml:"count"`
	Cwd      string              `yaml:"cwd"`
	Command  string              `yaml:"command"`
	Hooks    map[string][]string `yaml:"hooks"`
	Environs map[string]string   `yaml:"environs"`
}

func (service *Service) withOsEnvirons() []string {
	environs := os.Environ()
	for key, val := range service.Environs {
		environs = append(environs, fmt.Sprintf("%s=%s", key, val))
	}
	return environs
}

func (service *Service) expandedEnv() string {
	return os.ExpandEnv(os.Expand(service.Cwd, func(key string) string {
		if env, found := service.Environs[key]; found {
			return env
		}
		return fmt.Sprintf("${%s}", key)
	}))
}

func (service *Service) parsedCommand() (*syntax.File, error) {
	cmd, err := syntax.NewParser().Parse(strings.NewReader(service.Command), "")
	return cmd, err
}

// ParseConfigFile is a function that parses the given `filepath` configs.
func ParseConfigFile(filePath string) (map[string]*Service, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	raw := map[string]interface{}{}
	err = yaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	return parseServices(raw["services"])
}

func parseServices(servicesMap interface{}) (map[string]*Service, error) {
	raw, err := yaml.Marshal(servicesMap)
	if err != nil {
		return nil, err
	}

	services := make(map[string]*Service)
	err = yaml.Unmarshal(raw, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}
