package processor

import (
	"vec/config"
	"vec/model"
)

type Processors []Processor

func NewProcessors() Processors {
	return make(Processors, 0)
}

func (ps Processors) Add(processor Processor) Processors {
	ps = append(ps, processor)
	return ps
}

type message struct {
	vector
	vectorID int
	*model.Patent
}

func (ps Processors) Process(pchan chan *model.Patent) {
	chans := fanout(pchan, len(ps), config.Get().ConcurrencyConfig.PatentPoolSize)
	stopChan := make(chan struct{})
	for i, p := range ps {
		processOneField(p, chans[i], stopChan)
	}
	for i := 0; i < 3*len(ps); i++ {
		<-stopChan
	}
}

func processOneField(p Processor, pchan chan *model.Patent, stopChan chan struct{}) {
	vectorChanSize := config.Get().ConcurrencyConfig.VectorPoolSize
	mchan := make(chan *message, vectorChanSize)
	go genVec(p, pchan, mchan, stopChan)
	mchans := fanout(mchan, 2, vectorChanSize)
	go saveVec(p, mchans[0], stopChan)
	go saveVecIDAndPatentID(p, mchans[1], stopChan)
}

func fanout[T any](originChan chan T, num, size int) []chan T {
	chans := make([]chan T, num)
	for i := 0; i < num; i++ {
		chans[i] = make(chan T, size)
	}
	go func() {
		for v := range originChan {
			for _, ch := range chans {
				ch <- v
			}
		}
		defer func() {
			for _, ch := range chans {
				close(ch)
			}
		}()
	}()
	return chans
}

// 并行地将专利转换为向量
func genVec(p Processor, pchan chan *model.Patent, mchan chan *message, stop chan<- struct{}) {
	i := 0
	for patent := range pchan {
		vectorID := i
		m := &message{
			vector:   p.ToVec(patent),
			vectorID: vectorID,
			Patent:   patent,
		}
		mchan <- m
		i++
	}
	defer func() {
		close(mchan)
		stop <- struct{}{}
	}()
}

// 并行地将向量存储到文件
func saveVec(p Processor, mchan chan *message, stop chan<- struct{}) {
	for m := range mchan {
		err := p.SaveVec(m.vector)
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		stop <- struct{}{}
	}()
}

// 并行地将向量ID和专利ID的对应关系存储到数据库
func saveVecIDAndPatentID(p Processor, mchan chan *message, stop chan<- struct{}) {
	for m := range mchan {
		err := p.SaveVecIDAndPatentID(m.Patent, m.vectorID)
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		stop <- struct{}{}
	}()
}
