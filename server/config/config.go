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
	_,err := utils.DeSerializeData(configData, &Config)
	if err!= nil {
		panic("序列化配置失败：" + err.Error())
	}
	log.Printf("success load config from %s",utils.PrintDataAsJson(Config))
}
