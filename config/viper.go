package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	config := viper.New()
	// Ubah path ini sesuai path lokal Anda
	config.SetConfigFile("/home/wedha/Documents/uasgo/config.json") 
	err := config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	return config
}