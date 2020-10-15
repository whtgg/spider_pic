package utils

import (
	"encoding/base64"
	"os"
	"time"
)

/**
 * 生成新的文件名
 */
func base64String(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

/***
 * 新图片的绝对路径
*/
func newImagePath(name string) string{
	return createDir()+"/"+name+".jpeg"
}

/**
 * 生成新的文件目录
*/
func createDir() string{
	childDir := time.Now().Format("2006-01-02")
	cDir,_ := os.Getwd()
	imageDir := cDir+"/image/"+childDir
	if !checkExist(imageDir) {
		os.MkdirAll(imageDir,os.ModePerm)
		os.Chmod(imageDir,os.ModePerm)
		return imageDir
	}
	return imageDir
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}


