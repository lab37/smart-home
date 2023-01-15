package main

import (
	"image"
	"log"
	"os/exec"
	"time"
	"flag"
	"github.com/lucacasonato/mqtt"
)

func main() {
	var configFilePath string
	flag.StringVar(&configFilePath, "config", `F:\Program Files\new-face-detect\config.json`, "配置文件路径")
	flag.Parse()

	// Config global
	var Config = loadConfig(configFilePath)
	mqttClient := newMqttClient(Config.MQTTserver, Config.MQTTuserName, Config.MQTTpassword, "face-recognition")
	// 订阅主题
	mqttSubWithTimeout(mqttClient, Config.MQTTsubTopic, 1*time.Second)

	// 设置对应主题的处理函数
	mqttClient.Handle(Config.MQTTsubTopic, func(m mqtt.Message) {
		log.Println("门口有人移动，准备推送门口摄像机视频信号")
		cmd := exec.Command(Config.FFmpegScriptFile)
		cmd.Run()

		/*
			@echo off
			tasklist /nh|findstr /i "ffmpeg.exe"
			if ERRORLEVEL 1 (F:\tools\ffmpeg\ffmpeg -i "rtsp://192.168.31.96:8554/gate" -y -f image2 -r 4/1 -update 1   -s 960x540  -vf format=gray  -t  30  F:\tools\ffmpeg\rtsp.jpg) else (echo "ffmpeg already exist")
		*/
		// 推流命令执行以后发送MQTT消息通知
		/* 		err := mqttClient.PublishString(ctx(), "security/gate/rtsp/start", "ok", mqtt.AtLeastOnce)
		   		if err != nil {
		   			log.Printf("failed to publish: %v\n", err)
		   		}
		*/

	})

	// 生成人脸检测器
	faceDetectClassifier := getFaceDetectClassifier(Config.FaceFinder)

	// 加载已知人脸识别数据库
	faceDescriptions, names := loadFacesDatabase(Config.FaceData)

	// 生成对应模型的人脸识别器
	faceRecogizer := getFaceRecognizer(Config.TestDataDir, faceDescriptions)
	defer faceRecogizer.Close()

	// 建立人名传递通道
	nameQueue := make(chan string, 20)
	// 建立图片传递通道
	imgQueue := make(chan image.Image, 30)
	// 开协程收集图片
	go monitAndPutNewImgToChan(Config.ImgFileName, imgQueue)
	// 开协识别图片
	go func() {
		log.Println("now start detect goroute")

		for tmpImg := range imgQueue {
			if len(imgQueue) > 10 {
				log.Println("图片识别不过来，积压了！, 当前已积压：", len(imgQueue), "张图片。你能换个CPU吗!")
			}
			numberOfFace := detectFace(faceDetectClassifier, tmpImg)
			if numberOfFace > 0 {
				recognizeFaceAndPushName(faceRecogizer, names, tmpImg, nameQueue)
			}
		}

	}()
	go countAndPublicName(nameQueue, mqttClient, Config.MQTTpubTopic)
	select {}
}
