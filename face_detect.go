package main

import (
	"image"
	"log"
	"os"
	"time"

	//	"time"

	pigo "github.com/esimov/pigo/core"
)

// 使用特定的模型对象来检测图片中的人脸。计算出人脸在图片的位置与置信度，返回人脸个数。
func detectFace(classifier *pigo.Pigo, img image.Image) (numberOfFace int) {
	numberOfFace = 0
	start := time.Now()
	angle := 0.0 // cascade rotation angle. 0.0 is 0 radians and 1.0 is 2*pi radians

	src := pigo.ImgToNRGBA(img)

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     200,
		MaxSize:     540,
		ShiftFactor: 0.1,
		ScaleFactor: 1.4,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}
	// 下面的注释是原厂带的，不要问我为什么没有中文注释，因为我英语暂时还没过四级。
	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cParams, angle)
	var qTresh float32 = 5
	goodQ := []pigo.Detection{}
	// 筛选出得分大于6.8的结果,放入goodQ中
	for i := range dets {
		if dets[i].Q > qTresh {
			goodQ = append(goodQ, dets[i])
		}
	}
	if len(goodQ) > 0 {
		// 根据各个人脸区域的交并比(IoU), 鉴别出有几张人脸.
		dets2 := classifier.ClusterDetections(goodQ, 0.2)
		if len(dets2) > 0 {
			numberOfFace=len(dets2)
			log.Println("来了", len(dets2), "个人")
			elapsed := time.Since(start)
			log.Printf("本次检测耗时: %s", elapsed)
		}
	}
	//elapsed := time.Since(start)
	//log.Printf("本次检测耗时: %s", elapsed)
	return
}

// 解析模型文件，返回解析后得到的模型对象。
func getFaceDetectClassifier(modelPath string) (classifier *pigo.Pigo) {
	cascadeFile, err := os.ReadFile(modelPath)
	if err != nil {
		log.Fatalln("什么狗FaceDetect模型路径, 读不了: ", err)
	}

	p := pigo.NewPigo()

	// 下面的注释也是原厂带的，但这个我好像认识不少单词，只是整句话的意思不认识。
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err = p.Unpack(cascadeFile)
	if err != nil {
		log.Fatalln("路径倒是没问题了，但这熊文件解析不了啊: ", err)
	}
	return
}
