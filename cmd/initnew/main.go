package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/reugn/go-streams"
	"github.com/reugn/go-streams/flow"

	"vec/config"
	"vec/db"
	"vec/diskann"
	"vec/model"
	"vec/model/vector"
	"vec/processor"
	streamUtil "vec/stream"

	ext "github.com/reugn/go-streams/extension"
)

func main() {
	config.Init()
	db.Init()
	vector.Init()

	database := db.Get()
	fmt.Println("数据库初始化成功")
	fmt.Println(database)

	patentChanSize := config.Get().ConcurrencyConfig.PatentPoolSize
	patents := make(chan *model.Patent, patentChanSize)
	patentsTypeAny := make(chan any, patentChanSize)

	go transformChan(patents, patentsTypeAny)

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

	source := ext.NewChanSource(patentsTypeAny)
	flows := flow.FanOut(source, len(processors))

	//sink := ext.NewStdoutSink()

	var wg sync.WaitGroup

	for i, p := range processors {
		fmt.Printf("flows %d: in:%v, out:%v\n", i, flows[i].In(), flows[i].Out())
		p := p
		// 把专利分批，统一请求向量化服务（每批500个，目前向量化接口一批最多能容纳510个）
		batchFlow := streamUtil.NewBatchFlow(500, 1*time.Second)
		// 获取转成向量后的flow
		// 最终的结果是，每个元素都是vectorPatentAndVectorID，即 包含专利、向量、向量id的结构体
		vectorAndPatentFlow := flows[i].
			Via(batchFlow).        // 每n个转成一批
			Via(strToVecBatch(p)). // 批量转成向量（返回还是数组）
			Via(flatten(20)).      // 扁平化
			Via(addVectorID(IDGenerator()))

		// fan out 一下，分别给向量文件保存和redis存储用
		messageFlows := flow.FanOut(vectorAndPatentFlow, 2)

		saveVecFlow := messageFlows[0]
		//saveVecSink := ext.NewFileSink(vector.GetIndexVectorFullPath(p.Field()))
		saveVecSink := streamUtil.NewBinaryFileSink(vector.GetIndexVectorFullPath(p.Field()))

		redisFlow := messageFlows[1]
		//redisSink := ext.NewFileSink(vector.GetIndexVectorFullPath(p.Field()) + ".redis")
		redisSink := streamUtil.NewRedisSink(p)

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			saveVecFlow.
				//Via(extractVector()).
				To(saveVecSink)
		}(&wg)

		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			redisFlow.
				//Via(mockSaveVecIDAndPatentID()).
				To(redisSink)
		}(&wg)
		fmt.Printf("生成处理器 %v 成功\n", p.Field())
	}

	// 这个stop收到了信号，说明专利查询协程已经结束了
	<-stop
	fmt.Println("111")

	wg.Wait() //等待前面的协程结束

	fmt.Println("str to vec done")

	// 测试
	fmt.Println("测试一下向量和专利的对应关系")
	testVecID := int64(99)

	patentID, err := model.GetPatentIDByVectorID("name", testVecID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("ID 为 <name-%d> 的向量对应的专利ID为 <%s>\n", testVecID, patentID)

	// 调用 diskann 建立索引
	diskann.BuildIndex()

	fmt.Println("所有任务完成（索引建立异步进行中）")

}

// 批量将专利转换成向量
func strToVecBatch(p processor.Processor) streams.Flow {
	f := func(patents []any) (ms []*streamUtil.VectorPatent) {
		patentsRealType := make([]*model.Patent, len(patents))
		for i, _ := range patents {
			patentsRealType[i] = patents[i].(*model.Patent)
		}

		now := time.Now()
		hashID := getHashValue(now.Format("2006-01-02 15:04:05") + p.Field())
		fmt.Printf("%v, 转换专利到向量，本批专利数量：%d, 向量化字段：%v，本次任务标识：%v\n",
			now.Format("2006-01-02 15:04:05"), len(patentsRealType), p.Field(), hashID)

		vectors := p.ToVecs(patentsRealType)

		for i, patent := range patentsRealType {
			ms = append(ms, &streamUtil.VectorPatent{
				Vector: vectors[i],
				Patent: patent,
			})
		}
		fmt.Printf("%v, 转换专利到向量，本批专利数量：%d, 向量化字段：%v, 本次任务标识：%v，转换完成\n",
			time.Now().Format("2006-01-02 15:04:05"), len(patentsRealType), p.Field(), hashID)

		return ms
	}

	return flow.NewMap[[]any, []*streamUtil.VectorPatent](f, 20)
}

// 把一组专利拉平成一个个
func flatten(parallelism uint) streams.Flow {
	return flow.NewFlatMap(func(element []*streamUtil.VectorPatent) []*streamUtil.VectorPatent {
		return element
	}, parallelism)
}

// 为向量添加向量id
func addVectorID(getID func() int64) streams.Flow {
	f := func(m *streamUtil.VectorPatent) *streamUtil.VectorPatentAndVectorID {
		return &streamUtil.VectorPatentAndVectorID{VectorPatent: m, VectorID: getID()}
	}

	return flow.NewMap[*streamUtil.VectorPatent, *streamUtil.VectorPatentAndVectorID](f, 20)
}

// 提取向量（纯测试用）
func extractVector() streams.Flow {
	f := func(m *streamUtil.VectorPatentAndVectorID) string {
		return m.Vector.Describe() + "\n"
	}

	// 这里写文件要保证向量的顺序，所以并发为1
	return flow.NewMap[*streamUtil.VectorPatentAndVectorID, string](f, 1)
}

//// 保存向量成文件
//func saveVector(p processor.Processor) streams.Flow {
//	f := func(m *VectorPatentAndVectorID) {
//
//	}
//}

// mock 一下保存向量id和专利id的流程
func mockSaveVecIDAndPatentID() streams.Flow {
	f := func(m *streamUtil.VectorPatentAndVectorID) string {
		return fmt.Sprintf("专利<%s>对应的向量id是<%v>\n", m.Name, m.VectorID)
	}

	return flow.NewMap[*streamUtil.VectorPatentAndVectorID, string](f, 10)
}

// IDGenerator 从0开始生成ID，不断累加
func IDGenerator() func() int64 {
	var i int64 = -1
	return func() int64 {
		i++
		return i
	}
}

// 转换一下专利chan的类型，go-streams的输入是chan any
func transformChan(in <-chan *model.Patent, out chan<- any) {
	for e := range in {
		out <- e
	}
	close(out)
}

func getHashValue(str string) string {
	hashID := md5.New()
	hashID.Write([]byte(str))
	return hex.EncodeToString(hashID.Sum(nil))
}
