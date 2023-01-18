package main

import (
	"flag"
	"image"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/lucacasonato/mqtt"
)

var fileLogger *log.Logger

func main() {
	// 加载配置文件--------------------------------------------------------------
	var configFilePath string
	flag.StringVar(&configFilePath, "config", `F:\Program Files\new-face-detect\config.json`, "配置文件路径")
	flag.Parse()
	Config := loadConfig(configFilePath)

	// 建立日志记录工具---------------------------------------------------------
	f, err := os.OpenFile(Config.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Println("打开日志文件错误:", err)
		return
	}
	defer f.Close()
	logger := log.New(f, "face-detect: ", log.Lshortfile|log.Ldate|log.Ltime)
	fileLogger = logger

	// 生成mqtt客户端-----------------------------------------------------------
	mqttClient := newMqttClient(Config.MQTTserver, Config.MQTTuserName, Config.MQTTpassword, "face-recognition")
	// 订阅主题
	mqttSubWithTimeout(mqttClient, Config.MQTTsubTopic, 1*time.Second)
	// 设置对应主题的处理函数
	mqttClient.Handle(Config.MQTTsubTopic, func(m mqtt.Message) {
		cmd := exec.Command(Config.FFmpegScriptFile)
		cmd.Run()
	})

	// 生成人脸检测器-----------------------------------------------------------
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

	// 开协程收集图片---------------------------------------------------------
	go monitAndPutNewImgToChan(Config.ImgFileName, imgQueue)

	// 开协识别图片----------------------------------------------------------
	go func() {
		for tmpImg := range imgQueue {
			if len(imgQueue) > 10 {
				log.Println("图片识别不过来，积压了！, 当前已积压：", len(imgQueue), "张图片。你能换个CPU吗!")
				fileLogger.Println("图片识别不过来，积压了！, 当前已积压：", len(imgQueue), "张图片。")
			}
			numberOfFace := detectFace(faceDetectClassifier, tmpImg)
			if numberOfFace > 0 {
				// 这里必须用协程，不然会卡在这里等识别完才会去取下个图片
				go recognizeFaceAndPushName(faceRecogizer, names, tmpImg, nameQueue)
			}
		}
	}()

	// 开协程统计和报告人名------------------------------------------------
	go countAndPublicName(nameQueue, mqttClient, Config.MQTTpubTopic)

	// 监控进程管理信号---------------------------------------------------
	waitingSignal()
}
