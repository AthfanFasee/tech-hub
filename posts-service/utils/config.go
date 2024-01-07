package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	GrpcServerPort int    `mapstructure:"GRPC_SERVER_PORT"`
	PostgreDSN     string `mapstructure:"POSTGRESQL_DSN"`
	RabbitDSN      string `mapstructure:"RABBITMQ_DSN"`
}

var (
	configName = "app"
	configType = "env"
	configPath = "/"
)

// Read configuration from environemnt variables
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
