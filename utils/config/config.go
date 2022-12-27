package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

var AppPath string

type Boot struct {
	Token  string   `mapstructure:"token"`
	Admins []string `mapstructure:"admins"`
}

type Log struct {
	MaxSize    int `mapstructure:"max_size"`
	MaxAge     int `mapstructure:"max_age"`
	MaxBackups int `mapstructure:"max_backups"`
}

var BootC Boot
var LogC Log

func InitConfig() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	AppPath = path
	viper.SetConfigFile(path + "/config.toml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal("load config file err:", err)
	}
	err = viper.UnmarshalKey("log", &LogC)
	if err != nil {
		log.Fatal("load config log err:", err)
	}
	err = viper.UnmarshalKey("boot", &BootC)
	if err != nil {
		log.Fatal("load config log err:", err)
	}
}
