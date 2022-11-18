package main

import (
	"vec/config"
	"vec/db"
	"vec/processor"
	"vec/server"
)

// 目前这个服务有不少强绑定的东西
// 比如和 diskann 的索引建立、查询 服务的强绑定
func main() {
	config.Init()
	db.Init()
	processor.Init()
	server.Serve()
}
