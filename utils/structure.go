package utils

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

type RedisClient struct {
	Cmd 				*redis.Client
}

type QueryItem struct {
	ImgUrl 				string					`json:"thumbURL"`
}

type QueryResult struct {
	Data 				[]QueryItem				`json:"data"`
}

type Config struct {
	V						*viper.Viper
}

//任务
type Task struct {
	 Url							string
	 KeyWord						string
	 Width							int
	 Height							int
	 Image							string
	 Resize							string
}

type JsonResult struct {
	Code							int						`json:"code"`
	Msg								string					`json:"msg"`
}