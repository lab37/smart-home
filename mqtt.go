package main

import (
	"context"
	"log"
	"time"

	"github.com/lucacasonato/mqtt"
)

func newMqttClient(serverAddr string, username string, password string, clientID string) *mqtt.Client {
	mqttClient, err := mqtt.NewClient(mqtt.ClientOptions{
		// 必须项
		Servers: []string{
			serverAddr,
		},

		// 可选项
		ClientID:      clientID,
		Username:      username,
		Password:      password,
		AutoReconnect: true,
	})
	if err != nil {
		log.Println("无法创建mqtt客户端:", err)
		fileLogger.Println("无法创建mqtt客户端:", err)
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	//如果在1s内建立了连接则返回nil, 不然就是超时错误了.
	err = mqttClient.Connect(ctx)
	if err != nil {
		log.Println("连接mqttServer超时:", err)
		fileLogger.Println("连接mqttServer超时:", err)
		panic(err)
	}
	return mqttClient
}

func mqttPubWithTimeout(mqttClient *mqtt.Client, topic string, payload string, duration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	err := mqttClient.PublishString(ctx, topic, payload, mqtt.AtLeastOnce)
	if err != nil {
		log.Println("发布mqtt主题失败:", err)
		fileLogger.Println("发布mqtt主题失败:", err)
	}
}

func mqttSubWithTimeout(mqttClient *mqtt.Client, topic string, duration time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	err := mqttClient.Subscribe(ctx, topic, mqtt.AtMostOnce)
	if err != nil {
		log.Println("订阅mqtt主题失败:", err)
		fileLogger.Println("订阅mqtt主题失败:", err)
	}
}

// 下面这段只为了实时性这里采用监听绿米网关组播消息的方式来触发动作
/* 	multicastMessageCh := make(chan message, 100)
   	go udpMulticastReceiver("224.0.0.50", 9898, "", multicastMessageCh)
   	go func() {
   		for {
   			select {
   			case message := <-multicastMessageCh:
   				multicastData := dataST{}
   				multicastPayload := payload{}
   				json.Unmarshal(message.Data, &multicastData)
   				json.Unmarshal([]byte(multicastData.Payload), &multicastPayload)
   				if multicastPayload.RGB > 0 {
   					// 推流命令执行以后发送MQTT消息通知
   					err := client.PublishString(ctx(), "homeassistant/security/gate/motion", "ok", mqtt.AtLeastOnce)
   					if err != nil {
   						log.Printf("failed to publish: %v\n", err)
   					}
   				}
   			}
   		}
   	}() */

// 为不同主题添加处理函数
// 下面这个函数只用来在树莓派上订阅消息和推流的
/* 	client.Handle(`security/gate/motion`, func(m mqtt.Message) {
   		log.Println("门口有人移动，准备推送门口摄像机视频信号")

   		cmd := exec.Command("sh", "-c", Config.FFmpegScriptFile)
   		cmd.Run()

   		// 推流命令执行以后发送MQTT消息通知
   		err := client.PublishString(ctx(), "security/gate/rtsp/start", "ok", mqtt.AtLeastOnce)
   		if err != nil {
   			log.Printf("failed to publish: %v\n", err)
   		}
   	})
*/

// 下面这个处理函数只用于windows系统下弹窗提示有人来了
/* 	client.Handle(`security/gate/rtsp/start`, func(m mqtt.Message) {
	// log.Println("收到门口视频信号，准备播放")
	// time.Sleep(1 * time.Second)
	cmdline := `msg administrator /time:30  门口有人，已开启监控`
	cmd := exec.Command("cmd", "/c", "start "+cmdline)
	cmd.Run()
})
*/
