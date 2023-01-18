package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucacasonato/mqtt"
)

func countAndPublicName(nameCh chan string, mqttClient *mqtt.Client, mqttPubTopic string) {

	mqttTicker := time.NewTicker(time.Second * 2)
	defer mqttTicker.Stop()
	// 建立人名统计映射
	nameCount := make(map[string]int)
	for {
		select {
		case cName := <-nameCh:
			nameCount[cName] = nameCount[cName] + 1
		case <-mqttTicker.C:
			message := ""
			nums := 0
			cAnonymousNum := nameCount["anonymous"]
			for key, value := range nameCount {
				if value > 1 && key != "anonymous" {
					nums = nums + 1
					message = message + key + ","
				}
				nameCount[key] = 0
			}
			switch {
			case nums == 0:
				if cAnonymousNum > 2 {
					message = message + "有陌生人来了"
					log.Println(message)
					fileLogger.Println(message)
					mqttPubWithTimeout(mqttClient, mqttPubTopic, message, 1*time.Second)
				}
			case nums > 0:
				if cAnonymousNum > 2 {
					message = message + "来了, 带着陌生人"
					log.Println(message)
					fileLogger.Println(message)
					mqttPubWithTimeout(mqttClient, mqttPubTopic, message, 1*time.Second)
				} else {
					message = message + "来了"
					log.Println(message)
					fileLogger.Println(message)
					mqttPubWithTimeout(mqttClient, mqttPubTopic, message, 1*time.Second)
				}
			}
		}

	}
}

func waitingSignal() {
	c := make(chan os.Signal, 1)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT)

	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			fileLogger.Panicln("程序进程被结束...", s, ", 程序退出")
		default:
			fileLogger.Panicln("收到信号：", s, ", 程序退出")
		}
	}

}
