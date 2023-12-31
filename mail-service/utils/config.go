package util

import (
	"github.com/spf13/viper"
)

type Config struct {
	Domain         string `mapstructure:"MAILER_DOMAIN"`
	Host           string `mapstructure:"MAILER_HOST"`
	Port           string `mapstructure:"MAILER_PORT"`
	Username       string `mapstructure:"MAILER_USER_NAME"`
	Password       string `mapstructure:"MAILER_PASSWORD"`
	Encryption     string `mapstructure:"MAILER_ENCRYPTION"`
	FromName       string `mapstructure:"MAILER_FROM_NAME"`
	FromAddress    string `mapstructure:"MAILER_FROM_ADDRESS"`
	GrpcServerPort int    `mapstructure:"GRPC_SERVER_PORT"`
}

var (
	configName = "app"
	configType = "env"
	configPath = "."
)

// Read configuration from environemnt variables.
// Environemnt variables are replaced with correct values in prod environment.
func LoadEnv() (config Config, err error) {
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
