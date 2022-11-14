package main

import (
	"fmt"

	"vec/config"
	"vec/db"
	"vec/model"
	"vec/processor"
)

func main() {
	config.Init()
	db.Init()

	database := db.Get()
	fmt.Println("数据库初始化成功")
	fmt.Println(database)

	patentChanSize := config.Get().ConcurrencyConfig.PatentPoolSize
	patents := make(chan *model.Patent, patentChanSize)

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
		processors = processors.Add(processor.NewStrToVecMock(c.Field, c.Url))
	}

	fmt.Println("开始处理专利数据")
	processors.Process(patents)

	<-stop

	fmt.Println("str to vec done")

	fmt.Println("测试一下向量和专利的对应关系")
	testVecID := fmt.Sprintf("name-%d", 100)

	patentID, err := model.GetPatentIDByVectorID(testVecID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID 为 <%s> 的向量对应的专利ID为 <%s>", testVecID, patentID)
}
