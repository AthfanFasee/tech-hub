package utils

import (
	"github.com/spf13/viper"
)

type Config struct {
	GrpcServerPort int    `mapstructure:"GRPC_SERVER_PORT"`
	MongoURL       string `mapstructure:"MONGO_URL"`
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
