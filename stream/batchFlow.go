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
	done      chan struct{}
}

// Verify BatchFlow implements the streams.Flow interface
var _ streams.Flow = (*BatchFlow)(nil)

// NewBatchFlow returns a new BatchFlow instance
//
// it will batch the incoming elements into a slice
// whether the batch size is reached or the batch time is elapsed.
func NewBatchFlow(batchSize int, batchTime time.Duration) *BatchFlow {
	if batchSize <= 0 {
		panic("batch size should greater than 0")
	}

	bf := &BatchFlow{
		batchSize: batchSize,
		batchTime: batchTime,
		ticker:    time.NewTicker(batchTime),
		buffer:    make([]any, 0, batchSize),
		in:        make(chan any),
		out:       make(chan any),
		done:      make(chan struct{}),
	}
	go bf.batchBySize()
	go bf.batchByTime()

	return bf
}

// Via sends the flow to the next stage via specified flow
func (bf *BatchFlow) Via(flow streams.Flow) streams.Flow {
	go bf.transmit(flow)

	return flow
}

// To sends the flow to the given sink
func (bf *BatchFlow) To(sink streams.Sink) {
	bf.transmit(sink)
}

// Out returns the output channel of the BatchFlow
func (bf *BatchFlow) Out() <-chan any {
	return bf.out
}

// In returns the input channel of the BatchFlow
func (bf *BatchFlow) In() chan<- any {
	return bf.in
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
	close(bf.done)
	close(bf.out)
}

func (bf *BatchFlow) batchByTime() {
	defer bf.ticker.Stop()

	for {
		select {
		case <-bf.ticker.C:
			bf.flush()
		case <-bf.done:
			return
		}
	}
}

// flush sends the batched items to the next flow
func (bf *BatchFlow) flush() {
	bf.mu.Lock()
	buffer := bf.buffer
	bf.buffer = make([]any, 0, bf.batchSize)
	bf.ticker.Reset(bf.batchTime)
	bf.mu.Unlock()

	if len(buffer) > 0 {
		bf.out <- buffer
	}
}

func (bf *BatchFlow) transmit(inlet streams.Inlet) {
	for item := range bf.out {
		inlet.In() <- item
	}
	defer close(inlet.In())
}
