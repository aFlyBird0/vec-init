package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"vec/config"
)

func initMysql() *gorm.DB {
	db, err := CreateMysqlDB(struct {
		Addr           string
		User           string
		Pass           string
		DB             string
		ConnectTimeout uint
	}{Addr: config.Get().MysqlConfig.Host + ":" + fmt.Sprint(config.Get().MysqlConfig.Port),
		User:           config.Get().MysqlConfig.User,
		Pass:           config.Get().MysqlConfig.Password,
		DB:             config.Get().MysqlConfig.Database,
		ConnectTimeout: 10})
	if err != nil {
		panic("connect DB error: " + err.Error())
	}

	return db
}

func CreateMysqlDSN(dbInfo struct {
	Addr string
	User string
	Pass string
	DB   string
}) string {
	//user:password@/dbname?charset=utf8&parseTime=True&loc=Local
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbInfo.User, dbInfo.Pass, dbInfo.Addr, dbInfo.DB)
}

func CreateMysqlDB(dbInfo struct {
	Addr           string
	User           string
	Pass           string
	DB             string
	ConnectTimeout uint
}) (*gorm.DB, error) {
	cfg := struct {
		Addr string
		User string
		Pass string
		DB   string
	}{dbInfo.Addr, dbInfo.User, dbInfo.Pass, dbInfo.DB}
	DB, err := gorm.Open(mysql.Open(CreateMysqlDSN(cfg)), &gorm.Config{PrepareStmt: true})
	return DB, err
}
