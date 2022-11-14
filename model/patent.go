package model

import (
	"context"
	"fmt"

	"vec/config"
	"vec/db"
)

type Patent struct {
	ID            string
	Name          string
	ApplicationNo string
	AbstractCh    string
	Claim         string
}

func (p *Patent) TableName() string {
	return config.Get().MysqlConfig.Table
}

// GetField 根据字段名获取字段值
// todo 改成基于数据库DDL来
func (p *Patent) GetField(field string) string {
	switch field {
	case "name":
		return p.Name
	case "abstract":
		return p.AbstractCh
	case "claim":
		return p.Claim
	default:
		return ""
	}
}

func CountPatent() (int, error) {
	var count int64
	err := db.Get().Mysql.Model(&Patent{}).Count(&count).Error
	return int(count), err
}

func GetPatentList(pchan chan *Patent, stop chan<- struct{}) error {
	// get patents from database by page

	var (
		pageSize       = config.Get().ConcurrencyConfig.PageSize
		queryWorkerNum = config.Get().ConcurrencyConfig.QueryWorkerSize
	)

	pageChan := make(chan int, queryWorkerNum*2)            // 页数通道
	pageNotifyChan := make(chan struct{}, queryWorkerNum*2) // 通知页数通道递增页数
	// 控制每次只有一个查询能写入专利通道，而不是每个查询都能写入一部分
	// 这样能让某个查询快速把该页的专利写入，然后再去读取下一页
	queryWriteChan := make(chan struct{}, 1)
	stopQueryChan := make(chan struct{}, queryWorkerNum) // 停止查询通道,todo 用context代替

	ctx, cancel := context.WithCancel(context.Background())

	// 初始化
	for i := 0; i < queryWorkerNum; i++ {
		pageNotifyChan <- struct{}{}
	}
	queryWriteChan <- struct{}{}

	go generatePageChan(ctx, pageChan, pageNotifyChan)

	for i := 0; i < queryWorkerNum; i++ {
		ii := i
		// 因为后期查询会很慢，所以使用并发查询。使用 pchan 自动限速
		go func() {
			getPatentsOnePatch(ctx, pchan, pageChan, pageNotifyChan, stopQueryChan, queryWriteChan, ii, pageSize)
		}()
	}

	// 等待所有查询结束
	for i := 0; i < queryWorkerNum; i++ {
		<-stopQueryChan
		fmt.Println("查询结束", i)
	}
	// 取消所有协程
	cancel()

	close(stopQueryChan)
	close(pageNotifyChan)
	//close(pageChan)
	close(pchan)
	fmt.Println("关闭专利查询管道")

	stop <- struct{}{}

	return nil
}

func generatePageChan(ctx context.Context, pageChan chan int, pageNotifyChan chan struct{}) {
	page := 0
	for {
		select {
		case <-ctx.Done():
			close(pageChan)
			return
		case <-pageNotifyChan:
			pageChan <- page
			page++
		}
	}
}

func getPatentsOnePatch(ctx context.Context, pchan chan *Patent, pageChan chan int, pageNotify, stopQuery chan<- struct{}, queryWrite chan struct{}, workerIndex, pageSize int) {
	fmt.Printf("启动专利查询协程 %d, 页面大小 %d\n", workerIndex, pageSize)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("专利查询协程 %d 收到停止信号，退出\n", workerIndex)
			stopQuery <- struct{}{}
			return
		case page := <-pageChan:
			patents := make([]Patent, 0, pageSize)
			// mysql 计算总数太耗时了
			// 这里直接一直按页读下去，直到读到空为止
			// 从数据库中查询专利
			fmt.Printf("线程 %d 开始查询第 %d 页专利，页面大小为 %d\n", workerIndex, page, pageSize)
			err := db.Get().Mysql.Limit(pageSize).Offset(page * pageSize).Find(&patents).Error
			if err != nil {
				panic(err)
			}
			// 如果读到空，就退出
			if len(patents) == 0 {
				fmt.Printf("线程 %d 专利读取完毕\n", workerIndex)
				stopQuery <- struct{}{}
				return
			}

			// 如果专利非空
			pageNotify <- struct{}{}
			if page%100 == 0 {
				fmt.Printf("已经读取了 %d 页专利数据，每页 %d 条\n", page, pageSize)
			}

			// 将查询到的专利写入专利通道
			// 一个查询的专利写入完毕后，才能写入下一个查询的专利，提高效率
			// 获取写入专利的权限
			fmt.Printf("线程 %d 尝试获取写入专利的权限\n", workerIndex)
			<-queryWrite
			fmt.Printf("线程 %d 开始写入第 %d 页专利\n", workerIndex, page)
			for _, patent := range patents {
				patent := patent
				//fmt.Println("111", patent.Name, patent.ID, patent.Page)
				pchan <- &patent
			}
			// 释放写入专利的权限
			queryWrite <- struct{}{}
			fmt.Printf("线程 %d 写入第 %d 页专利完，释放写权限\n", workerIndex, page)
		default:
		}
	}

}
