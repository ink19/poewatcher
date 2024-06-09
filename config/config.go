package config

import (
	"log"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/ini"
	"github.com/gookit/config/v2/yaml"
)

type Config struct {
	Port   int `config:"port"`
	Notify struct {
		Type string `config:"wework"`
		URL  string `config:"url"`
	} `config:"notify"`
}

func init() {
	config.WithOptions(config.ParseEnv)
	config.WithOptions(config.WithTagName("config"))
	config.AddDriver(ini.Driver)
	config.AddDriver(yaml.Driver)
}

func Init(configFileName string) {
	if err := config.LoadFiles(configFileName); err != nil {
		log.Panicf("LoadFiles fail, err: %v", err)
	}
}
