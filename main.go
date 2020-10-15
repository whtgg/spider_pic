package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"picSearch/utils"
	"strconv"
	"time"
)

var (
	Conf *utils.Config
	config utils.Config
)

func fetch(res http.ResponseWriter,r *http.Request) {
	vars := mux.Vars(r)
	width,_ := strconv.Atoi(vars["width"])
	height,_ := strconv.Atoi(vars["height"])
	tasking := &utils.Task{
		KeyWord:vars["keyword"],
		Height: height,
		Width: width,
	}
	dst := tasking.Fetch()
	f,_ := os.Open(dst)
	img,_ := ioutil.ReadAll(f)
	expiration := time.Now()
	expiration = expiration.AddDate(0, 1, 0)
	res.Header().Set("Cache-Control","public,max-age=315360000")
	res.Header().Set("Expires",expiration.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	res.Write(img)
	return
}

func change(res http.ResponseWriter,r *http.Request) {
	vars := mux.Vars(r)
	width,_ := strconv.Atoi(vars["width"])
	height,_ := strconv.Atoi(vars["height"])
	tasking := &utils.Task{
		KeyWord:vars["keyword"],
		Height: height,
		Width: width,
	}
	dst := tasking.Change()
	f,_ := os.Open(dst)
	img,_ := ioutil.ReadAll(f)
	res.Write(img)
	return
}

func remove(res http.ResponseWriter,r *http.Request) {
	res.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	tasking := &utils.Task{
		KeyWord:vars["keyword"],
	}
	tasking.DeleteImage()
	jsn,_ := json.Marshal(utils.JsonResult{Code: 200,Msg: "ok"})
	res.Write(jsn)
	return
}

func HeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !utils.CheckWebsite(Conf,request.Host) {
			writer.WriteHeader(500)
			writer.Write([]byte("非法站点"))
			return
		}
		writer.Header().Set("Content-Type", "image/jpeg")
		next.ServeHTTP(writer, request)
	})
}

func main() {
	_,Conf = utils.LoadConfigYaml(&config)
	utils.InitRedis(fmt.Sprintf("%s",Conf.V.Get("redis.addr")),fmt.Sprintf("%s",Conf.V.Get("redis.password")))
	err := utils.WatchConfig(Conf)
	rx := mux.NewRouter()
	rx.Use(HeaderMiddleware)
	rx.HandleFunc("/{keyword}/{width}/{height}",fetch).Methods("GET")//不带尺寸
	rx.HandleFunc("/{keyword}/{width}/{height}/change",change).Methods("GET")
	rx.HandleFunc("/{keyword}/del",remove).Methods("GET")
	server := http.Server{
		Addr:":"+Conf.V.GetString("port"),
		ReadTimeout: time.Second*2,
		WriteTimeout: time.Second*2,
		Handler: rx,
	}
	if err = server.ListenAndServe(); err != nil {
		panic(err)
	}
}
