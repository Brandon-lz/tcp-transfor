package config

import (
	"log"

	"github.com/Brandon-lz/tcp-transfor/utils"
	"github.com/BurntSushi/toml"
)

var Config ConfigDF

type ConfigDF struct {
	Port int `toml:"port"`
	Timeout int `toml:"timeout"`
}

func LoadConfig() {
	var configData map[string]interface{}
	tomlFile := "config.toml"
	if _, err := toml.DecodeFile(tomlFile, &configData); err != nil {
		panic(err)
	}
	utils.DeSerializeData(configData, &Config)
	log.Printf("success load config from %s",utils.PrintDataAsJson(Config))
}
