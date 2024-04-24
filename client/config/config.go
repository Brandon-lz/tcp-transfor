package config

import (
	"github.com/Brandon-lz/tcp-transfor/utils"
	"github.com/BurntSushi/toml"
)

var Config ConfigDF

type ConfigDF struct {
	TargetAddr string `json:"target_addr" toml:"target_addr"`
	SourceAddr string `json:"source_addr" toml:"source_addr"`
} 

func LoadConfig() {
	var configData map[string]interface{}
	tomlFile := "config.toml"
	if _, err := toml.DecodeFile(tomlFile, &configData); err!= nil {
		panic(err)
	}
	utils.SerializeData(configData,&Config)
}