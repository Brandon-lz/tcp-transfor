package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

var Config ConfigDF

// config.toml example
// [server]
// host = "127.0.0.1:38080"
//
// [client]
// name = "ubuntu1"            # 全局唯一的客户端名称
//
// [[map]]
// local-port = 9090
// server-port = 9091

type ConfigDF struct {
	Server struct {
		Host string `json:"host"`
	} `json:"server"`
	Client struct {
		Name string `json:"name"`
	} `json:"client"`
	Map []struct {
		LocalHost string `json:"local-host"`
		LocalPort  int    `json:"local-port"`
		ServerPort int    `json:"server-port"`
	} `json:"map"`
}

// LoadConfig 从当前目录的 config.toml 读取配置；不存在时回退到仓库内 client/config.toml
func LoadConfig() error {
	candidates := []string{"config.toml", "../client/config.toml"}
	var file string
	var err error
	for _, c := range candidates {
		if _, statErr := os.Stat(c); statErr == nil {
			file = c
			break
		}
	}
	if file == "" {
		return fmt.Errorf("config.toml not found, tried: %v", candidates)
	}
	var configData map[string]interface{}
	if _, err = toml.DecodeFile(file, &configData); err != nil {
		return err
	}
	if _, err = utils.DeSerializeData(configData, &Config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", file, err)
	}
	return nil
}
