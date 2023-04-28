package handlers

import (
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port         int    `yaml:"port"`
	OpenAIAPIKey string `yaml:"openai_api_key"`
	Connectors   *sync.Map
}

func LoadConfig(filename string, config interface{}) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	str := os.ExpandEnv(string(b))
	err = yaml.Unmarshal([]byte(str), config)
	if err != nil {
		return err
	}
	return nil
}

func InitConfig(filename string) (*Config, error) {
	c := new(Config)
	err := LoadConfig(filename, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
