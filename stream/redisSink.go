package stream

import (
	"fmt"
	"os"

	"github.com/reugn/go-streams"

	"vec/db"
	"vec/model"
	"vec/model/vector"
	"vec/processor"
)

type RedisSink struct {
	p processor.Processor

	in chan any
}

var _ streams.Sink = (*RedisSink)(nil)

func NewRedisSink(p processor.Processor) *RedisSink {
	s := &RedisSink{
		p:  p,
		in: make(chan any),
	}

	go s.init()

	return s
}

const redisMaxRetry = 6

func (s *RedisSink) init() {
	for item := range s.in {
		v := item.(*VectorPatentAndVectorID)
		key := vector.AddFieldToVectorID(s.p.Field(), v.VectorID)
		value := v.Patent.ID
		for i := 0; i < redisMaxRetry; i++ {
			err := model.SetRedis(db.Get().Redis, key, value)
			if err != nil {
				fmt.Printf("set redis error, key:%v, patentID: %v, err: %v\n", key, value, err)
				if i == redisMaxRetry-1 {
					fmt.Printf("set redis reached max err time, key:%v, patentID: %v, err: %v\n", key, value, err)
					if err2 := saveFailedRedisKeyValue(key, value); err2 == nil {
						fmt.Printf("save failed redis kv successfully, key:%v, patentID: %v\n", key, value)
					} else {
						fmt.Printf("save failed redis kv failed, key:%v, patentID: %v\n", key, value)
					}
				}
				continue
			}
			break
		}
	}

}

func saveFailedRedisKeyValue(key, value string) error {
	file, err := os.OpenFile("init/fail.redis.kv.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%s,%s\n", key, value)
	if err != nil {
		return err
	}
	return nil
}

func (s *RedisSink) In() chan<- interface{} {
	return s.in
}
