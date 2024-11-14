package config

import "github.com/spf13/viper"

type configuration struct {
	MAPApiKey       string `mapstructure:"MAPS_API_KEY"`
	MAPClientID     string `mapstructure:"MAPS_CLIENT_ID"`
	MAPClientSecret string `mapstructure:"MAPS_CLIENT_SECRET"`
	BaseURL         string
}

var ConfigVars configuration

func LoadConfig() error {
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(&ConfigVars)
	if err != nil {
		return err
	}
	ConfigVars.BaseURL = "https://api.olamaps.io/places/v1"
	return nil
}
