package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// init 使用 ./application.yaml 初始化全局配置
func init() {
	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithField("config", "viper").WithError(err).Panicf("unable to read global config")
	}
}

var conf *config

func Init() {
	conf = readConfig()
}

func Get() *config {
	return conf
}

func readConfig() *config {
	c := &config{}
	err := viper.UnmarshalKey("mysql", &c.MysqlConfig)
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("redis", &c.RedisConfig)
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("str2vec", &c.Str2VecConfigs)
	if err != nil {
		panic(err)
	}
	err = viper.UnmarshalKey("concurrency", &c.ConcurrencyConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println(c.Str2VecConfigs)

	return c
}
