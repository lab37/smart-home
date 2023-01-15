package main

import (
	"encoding/json"
	"log"
	"os"
)

// ServerST struct
type ConfigST struct {
	ImgFileName      string `json:"imgFileName"`
	MQTTserver       string `json:"mqttServer"`
	MQTTuserName     string `json:"mqttUserName"`
	MQTTpassword     string `json:"mqttPassword"`
	FFmpegScriptFile string `json:"ffmpegScriptFile"`
	FaceFinder       string `json:"faceFinder"`
	FaceData         string `json:"faceData"`
	TestDataDir      string `json:"testDataDir"`
}

// 读取配置文件并生成附属结构
func loadConfig(configFilePath string) *ConfigST {
	var tmp ConfigST
	data, err := os.ReadFile(configFilePath)
	if err == nil {
		err = json.Unmarshal(data, &tmp)
		if err != nil {
			log.Fatalln(err)
		}

	} else {
		log.Println("读取配置文件失败:", err)
	}
	return &tmp
}
