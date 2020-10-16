/*
Yaml configuration file parsing utility.
*/

package utils

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

// Service is a struct that represents the service in the YAML file.
type Service struct {
	Cwd      string                 `yaml:"cwd"`
	Command  string                 `yaml:"command"`
	Order    []interface{}          `yaml:"order"`
	Hooks    map[string][]string    `yaml:"hooks"`
	Environs map[string]interface{} `yaml:"environs"`
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
