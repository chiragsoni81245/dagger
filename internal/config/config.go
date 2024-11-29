package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
	} `mapstructure:"database"`

	Server struct {
		Port     int    `mapstructure:"port"`
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"server"`

	App struct {
		Env   string `mapstructure:"env"`
		Debug bool   `mapstructure:"debug"`
	} `mapstructure:"app"`
}

func GetConfig() (*Config, error){
	// Set the path to the configuration file
	viper.SetConfigFile("config.yaml")

	// Read the configuration file
	err := viper.ReadInConfig()
	if err != nil {
        return nil, err
	}

	// Unmarshal the configuration into a struct
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
        return nil, err
	}

    return &config, nil
}
