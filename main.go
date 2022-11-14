package main

import (
	"fmt"

	"vec/config"
	"vec/db"
	"vec/model"
	"vec/processor"
)

func main() {
	database := db.Get()
	fmt.Println("数据库初始化成功")
	fmt.Println(database)

	const pageChanSize = 1000
	patents := make(chan *model.Patent, pageChanSize)

	stop := make(chan struct{})

	go func() {
		err := model.GetPatentList(patents, stop)
		if err != nil {
			panic(err)
		}
	}()

	processors := processor.NewProcessors()

	for _, c := range config.Get().Str2VecConfigs {
		processors.Add(processor.NewStrToVecMock(c.Field, c.Url))
	}

	processors.Process(patents)

	<-stop

	fmt.Println("str to vec done")

	fmt.Println("测试一下向量和专利的对应关系")

	testVecID := fmt.Sprintf("name-%d", 100)

	patentID, err := model.GetPatentIDByVectorID(testVecID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID 为 <%d> 的向量对应的专利ID为 <%s>", testVecID, patentID)
}
