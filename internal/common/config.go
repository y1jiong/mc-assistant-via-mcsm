package common

import (
	"log"
)

const (
	configFileName = "config.json"
)

type Config struct {
	ApiUrl       string `json:"api_url"`
	ApiKey       string `json:"api_key"`
	ServerName   string `json:"server_name"`
	DataFileName string `json:"data_file_name"`
}

func (c *Config) InitToFile() (err error) {
	*c = Config{
		ApiUrl:       "http://127.0.0.1:23333/api/execute",
		DataFileName: "data.json",
	}
	if err != nil {
		return
	}
	err = MarshalAndSave(c, configFileName)
	if err != nil {
		return
	}
	log.Println("初始化 " + configFileName + " 完成")
	return
}

func (c *Config) LoadFromFile() (err error) {
	return LoadAndUnmarshal(configFileName, c)
}
