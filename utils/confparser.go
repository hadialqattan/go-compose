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

type service struct {
	IgnoreFailures bool                `yaml:"isolated"`
	SubService     bool                `yaml:"subservice"`
	Count          int                 `yaml:"count"`
	Cwd            string              `yaml:"cwd"`
	Command        string              `yaml:"command"`
	Hooks          map[string][]string `yaml:"hooks"`
	Environs       map[string]string   `yaml:"environs"`
}

func (service *service) withOsEnvirons() []string {
	environs := os.Environ()
	for key, val := range service.Environs {
		environs = append(environs, fmt.Sprintf("%s=%s", key, val))
	}
	return environs
}

func (service *service) expandedEnv() string {
	return os.ExpandEnv(os.Expand(service.Cwd, func(key string) string {
		if env, found := service.Environs[key]; found {
			return env
		}
		return fmt.Sprintf("${%s}", key)
	}))
}

func (service *service) parsedCommand() (*syntax.File, error) {
	cmd, err := syntax.NewParser().Parse(strings.NewReader(service.Command), "")
	return cmd, err
}

func parseConfigFile(filePath string) (map[string]*service, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	raw := make(map[string]interface{})
	err = yaml.Unmarshal(data, &raw)
	if err != nil {
		return nil, err
	}

	return parseServices(raw["services"])
}

func parseServices(servicesMap interface{}) (map[string]*service, error) {
	raw, err := yaml.Marshal(servicesMap)
	if err != nil {
		return nil, err
	}

	services := make(map[string]*service)
	err = yaml.Unmarshal(raw, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}
