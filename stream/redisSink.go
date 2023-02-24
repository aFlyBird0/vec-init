package stream

import (
	"fmt"

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

func (s *RedisSink) init() {
	for v := range s.in {
		value := v.(*VectorPatentAndVectorID)
		key := vector.AddFieldToVectorID(s.p.Field(), value.VectorID)
		if err := model.SetRedis(db.Get().Redis, key, value.Patent.ID); err != nil {
			fmt.Printf("set redis error: %v\n", err)
		}
	}

}

func (s *RedisSink) In() chan<- interface{} {
	return s.in
}
