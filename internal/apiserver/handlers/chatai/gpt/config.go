package gpt

type Config struct {
	Token string `yaml:"token" json:"token"  validate:"required"`
}
