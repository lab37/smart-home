package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func monitAndPutNewImgToChan(fileName string, imgQueue chan image.Image) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for {
		_, err := os.ReadFile(fileName)
		if err == nil {
			break
		}
		log.Println("等待图片生成中.......")
		time.Sleep(time.Duration(1) * time.Second)
	}
	err = watcher.Add(fileName)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("FileMonitor err", err)
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				imgBytes, err := os.ReadFile(fileName)
				if err != nil {
					continue
				}
				if len(imgBytes) == 0 {
					continue
				}
				img, err := jpeg.Decode(bytes.NewReader(imgBytes))
				if err != nil {
					continue
				}
				if len(imgQueue) < cap(imgQueue) {
					imgQueue <- img
				} else {
					log.Println("图片通道满了,已经", len(imgQueue), "张没处理了,无法往里放入新图片了")
				}

			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

}
