package ernie

type Config struct {
	AccessKey string `json:"access_key" yaml:"access_key" validate:"required"`
	SecretKey string `json:"secret_key" yaml:"secret_key" validate:"required"`
}
