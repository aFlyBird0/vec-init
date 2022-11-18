package main

import (
	"fmt"

	"vec/config"
	"vec/db"
	"vec/model"
	"vec/processor"
)

// todo 所有的错误处理
// todo 日志

func main() {
	config.Init()
	db.Init()
	processor.Init()

	database := db.Get()
	fmt.Println("数据库初始化成功")
	fmt.Println(database)

	patentChanSize := config.Get().ConcurrencyConfig.PatentPoolSize
	patents := make(chan *model.Patent, patentChanSize)

	// 停止信号
	stop := make(chan struct{})

	// 启动专利查询协程
	go func() {
		err := model.GetPatentList(patents, stop)
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("开始生成处理器")
	processors := processor.NewProcessors()
	for _, c := range config.Get().Str2VecConfigs {
		processors = processors.Add(processor.NewStrToVec(c.Field, c.Url))
	}

	// 专利处理（向量化、fvecs存储、对应关系存储）
	fmt.Println("开始处理专利数据")
	processors.Process(patents)

	<-stop

	fmt.Println("str to vec done")

	// 测试
	fmt.Println("测试一下向量和专利的对应关系")
	testVecID := "100"

	patentID, err := model.GetPatentIDByVectorID("name", testVecID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID 为 <name-%s> 的向量对应的专利ID为 <%s>", testVecID, patentID)
}
