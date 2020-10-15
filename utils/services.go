package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BASEURL = "https://image.baidu.com/search/acjson?ipn=rj&latest=&oe=utf-8&pn=1&rn=20&z=9&tn=resultjson_com&word="
)


func (t *Task) resize() (err error,resizePath string){
	fmt.Println("resize",*t)
	image,_ := imaging.Open(t.Image)
	image = imaging.Resize(image,t.Width,t.Height,imaging.Lanczos)
	originPath := Redis.Cmd.HGet(base64String(t.KeyWord),"origin").Val()
	originPathSplit := strings.Split(originPath,".")
	resizePath = originPathSplit[0]+"@"+strconv.Itoa(t.Width)+"X"+strconv.Itoa(t.Height)+".jpeg"
	err = imaging.Save(image,resizePath)
	Redis.Cmd.HSet(base64String(t.KeyWord),strconv.Itoa(t.Width)+"X"+strconv.Itoa(t.Height),resizePath)//更新
	return imaging.Save(image,resizePath),resizePath
}

func (t *Task) loadImage() {
	fmt.Println("loadImage",*t)
	req,err := http.NewRequest("GET",t.Url,nil)
	if err != nil {
		log.Printf("image request %s",err.Error())
		return
	}
	req.Header.Add("referer",BASEURL)
	client := &http.Client{}
	res,err := client.Do(req)
	if err != nil {
		log.Printf("client Do %s",err.Error())
		return
	}
	defer res.Body.Close()
	imgPath := newImagePath(base64String(time.Now().Format("2006-01-02 15:04:05")))
	file,err := os.Create(imgPath)
	if err != nil {
		log.Printf("os create %s",err.Error())
		return
	}
	_,err = io.Copy(file,res.Body)
	if err != nil {
		log.Printf("copy fail %s",err.Error())
		return
	}
	keywordBase64 := base64String(t.KeyWord)
	Redis.Cmd.SAdd("images",keywordBase64)
	Redis.Cmd.HSet(keywordBase64,"origin",imgPath)
	t.Image = imgPath
}

func (t *Task) loadHtml(){
	fmt.Println("loadHtml",*t)
	parseUrl,err := url.Parse(BASEURL)
	if err != nil {
		log.Printf("url parse err %s",err.Error())
		return
	}
	urlValues, err := url.ParseQuery(parseUrl.RawQuery)
	if err != nil {
		log.Printf("url ParseQuery err %s",err.Error())
		return
	}
	urlValues.Add("word",t.KeyWord)
	parseUrl.RawQuery = urlValues.Encode()
	resp,err := http.Get(fmt.Sprintf("%s",parseUrl))
	if err != nil {
		log.Printf("http get err %s",err.Error())
		return
	}
	defer resp.Body.Close()
	resBytes,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read bytes err %s",err.Error())
		return
	}
	var results QueryResult
	_ = json.Unmarshal(resBytes,&results)
	t.Url = results.Data[0].ImgUrl//更新值
	results.Data = results.Data[1:]
	wait,_ := json.Marshal(results)
	Redis.Cmd.HSet(base64String(t.KeyWord),"waiting",string(wait))
	return
}

func (t *Task) Fetch() string {
	var (
		path			string
	)
	keywordKey := base64.StdEncoding.EncodeToString([]byte(t.KeyWord))
	//已经请求过
	if Redis.Cmd.SIsMember("images",keywordKey).Val() {
		path = Redis.Cmd.HGet(keywordKey, "origin").Val()
		if t.Height > 0 && t.Width > 0 {
			path = Redis.Cmd.HGet(keywordKey, strconv.Itoa(t.Width)+"X"+strconv.Itoa(t.Height)).Val()
		}
		if len(path) > 0 {
			return path
		}
		//如果没有尺寸则修改原尺寸
		t.Image = Redis.Cmd.HGet(base64String(t.KeyWord),"origin").Val()
		_,p := t.resize()
		return p
	}
	//未请求过
	t.loadHtml()
	t.loadImage()
	_,p := t.resize()
	return p
}

func (t *Task) Change() string{
	var results QueryResult
	waitingImage := Redis.Cmd.HGet(base64String(t.KeyWord),"waiting").Val()
	if len(waitingImage) < 10 {
		t.loadHtml()
		t.loadImage()
		_,p := t.resize()
		return p
	}
	if err := json.Unmarshal([]byte(waitingImage),&results); err != nil {
		log.Printf("Unmarshal err %s",err.Error())
		return ""
	}
	//没有预备 重新请求
	if len(results.Data) < 2 {
		t.loadHtml()
		t.loadImage()
		_,p := t.resize()
		return p
	}
	t.Url = results.Data[0].ImgUrl//更新值
	t.loadImage()
	_,p := t.resize()
	results.Data = results.Data[1:]
	wait,_ := json.Marshal(results)
	Redis.Cmd.HSet(base64String(t.KeyWord),"waiting",string(wait))
	return p
}

/**
* 判断是否在哪个网站中
*/
func CheckWebsite(conf *Config,req string) bool{
	webs := conf.V.GetStringSlice("website")
	la := false
	for _,v := range webs {
		if v == req {
			la = true
			break
		}
	}
	return la
}


/***
 * 删除根据描述信息的图片
*/
func (t *Task) DeleteImage() {
	base64Key := base64String(t.KeyWord)
	r,_ := Redis.Cmd.HGetAll(base64Key).Result()
	for k,v := range r {
		if k != "waiting" {
			os.Remove(v)
		}
	}
	Redis.Cmd.Del(base64String(t.KeyWord))
	Redis.Cmd.SRem("images",base64Key)
}

