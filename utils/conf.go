package utils

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
)

func LoadConfigYaml(c *Config)  (error,*Config){
	path,_ := os.Getwd()
	c.V = viper.New()
	c.V.SetConfigName("config")
	c.V.AddConfigPath(path)
	c.V.AddConfigPath("./conf/")
	c.V.SetConfigType("yaml")
	if err := c.V.ReadInConfig(); err != nil{
		return  err,nil
	}
	return nil,c
}

func WatchConfig(c *Config) error {
	if err,_ := LoadConfigYaml(c); err !=nil{
		return err
	}
	c.V.WatchConfig()
	watch := func(e fsnotify.Event) {
		log.Printf("Config file is changed: %s \n", e.String())
	}
	c.V.OnConfigChange(watch)
	return nil
}

