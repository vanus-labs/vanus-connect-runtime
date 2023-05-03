package handlers

import (
	"os"

	"gopkg.in/yaml.v2"

	"github.com/vanus-labs/vanus-connect-runtime/pkg/controller"
)

type Config struct {
	Port         int    `yaml:"port"`
	OpenAIAPIKey string `yaml:"openai_api_key"`
	Ctrl         *controller.Controller
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
