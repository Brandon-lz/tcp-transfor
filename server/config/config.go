package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/Brandon-lz/tcp-transfor/utils"
)

var Config ConfigDF

type ConfigDF struct {
	Port    int `toml:"port" json:"port"`
	Timeout int `toml:"timeout" json:"timeout"`
	UIPort  int `toml:"ui-port" json:"ui-port"` // 只读状态接口端口，只监听 127.0.0.1
}

// LoadConfig 从当前目录的 config.toml 读取配置；不存在时回退到仓库内 server/config.toml
func LoadConfig() {
	candidates := []string{"config.toml", "../server/config.toml"}
	var file string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			file = c
			break
		}
	}
	if file == "" {
		panic("config.toml not found")
	}

	var configData map[string]interface{}
	if _, err := toml.DecodeFile(file, &configData); err != nil {
		panic(err)
	}
	_, err := utils.DeSerializeData(configData, &Config)
	if err != nil {
		panic("序列化配置失败：" + err.Error())
	}
	log.Printf("success load config from %s: %s", file, utils.PrintDataAsJson(Config))
}
