package stream

import (
	"sync"
	"time"

	"github.com/reugn/go-streams"
)

type BatchFlow struct {
	batchSize int
	batchTime time.Duration
	ticker    *time.Ticker
	in        chan any
	out       chan any
	buffer    []any
	mu        sync.Mutex
}

// NewBatchFlow returns a new BatchFlow instance
func NewBatchFlow(batchSize int, batchTime time.Duration) *BatchFlow {
	bf := &BatchFlow{
		batchSize: batchSize,
		batchTime: batchTime,
		ticker:    time.NewTicker(batchTime),
		buffer:    make([]any, 0, batchSize),
		in:        make(chan any),
		out:       make(chan any, 20),
	}
	go bf.batchBySize()
	go bf.batchByTime()

	return bf
}

// In returns the input channel of the BatchFlow
func (bf *BatchFlow) In() chan<- any {
	return bf.in
}

// Out returns the output channel of the BatchFlow
func (bf *BatchFlow) Out() <-chan any {
	return bf.out
}

func (bf *BatchFlow) transmit(inlet streams.Inlet) {
	for item := range bf.out {
		inlet.In() <- item
	}
	defer close(inlet.In())

}

// Via sends the flow to the next stage via specified flow
func (bf *BatchFlow) Via(flow streams.Flow) streams.Flow {
	go bf.transmit(flow)

	return flow
}

func (bf *BatchFlow) batchBySize() {
	for elem := range bf.in {
		bf.mu.Lock()
		bf.buffer = append(bf.buffer, elem)
		bf.mu.Unlock()
		if len(bf.buffer) >= bf.batchSize {
			bf.flush()
		}
	}
	close(bf.out)
}

func (bf *BatchFlow) batchByTime() {
	for {
		select {
		case <-bf.ticker.C:
			bf.flush()
		}
	}
}

// To sends the flow to the given sink
func (bf *BatchFlow) To(sink streams.Sink) {
	bf.transmit(sink)
}

// flush sends the batched items to the next flow
func (bf *BatchFlow) flush() {
	bf.mu.Lock()
	buffer := bf.buffer
	bf.buffer = make([]any, 0, bf.batchSize)
	bf.mu.Unlock()

	if len(buffer) > 0 {
		bf.out <- buffer
	}
}

// Ensure that BatchFlow implements the streams.Flow interface
var _ streams.Flow = (*BatchFlow)(nil)
