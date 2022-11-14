package model

import (
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

	const pageSize = 1000

	// mysql 计算总数太耗时了
	// 这里直接一直读下去，直到读到空为止

	page := 0

	for {
		patents := make([]Patent, 0, pageSize)
		// todo 因为后期查询会很慢，所以改成并发查询。使用 channel 自动限速
		err := db.Get().Mysql.Limit(pageSize).Offset(page * pageSize).Find(&patents).Error
		if err != nil {
			return err
		}
		if len(patents) == 0 {
			break
		}
		for _, patent := range patents {
			patent := patent
			//fmt.Println("111", patent.Name, patent.ID, patent.Page)
			pchan <- &patent
		}
		page++
		if page%100 == 0 {
			fmt.Printf("已经读取了 %d 页专利数据，每页 %d 条\n", page, pageSize)
		}
	}

	close(pchan)
	fmt.Println("关闭专利查询管道")
	stop <- struct{}{}

	return nil
}
