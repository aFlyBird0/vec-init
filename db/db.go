package db

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type DBList struct {
	Mysql *gorm.DB
	Redis *redis.Client
}

var db *DBList

func Get() *DBList {
	if db == nil {
		db = new(DBList)
		db.Mysql = initMysql()
		db.Redis = initRedis()
	}
	return db
}
