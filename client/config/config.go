package config

import (
	"github.com/Brandon-lz/tcp-transfor/utils"
	"github.com/BurntSushi/toml"
)

var Config ConfigDF

// config.toml
// [server]
// host = "127.0.0.1:8080"

// [client]
// name = "ubuntu1"            # 全局唯一的客户端名称

// [[map]]
// local-port = 9090
// server-port = 9090

// [[map]]
// local-port = 9091
// server-port = 9091

type ConfigDF struct {
	Server struct {
		Host string `json:"host"`
	} `json:"server"`
	Client struct {
		Name string `json:"name"`
	} `json:"client"`
	Map []struct {
		LocalPort  int `json:"local-port"`
		ServerPort int `json:"server-port"`
	} `json:"map"`
}

func LoadConfig() error {
	var configData map[string]interface{}
	tomlFile := "config.toml"
	if _, err := toml.DecodeFile(tomlFile, &configData); err != nil {
		return err
	}
	utils.DeSerializeData(configData, &Config)
	return nil
}
