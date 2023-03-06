package stream

import (
	"sync"
	"time"

	"github.com/reugn/go-streams"
)

type Batch[T any] struct {
	batchSize uint
	batchTime time.Duration
	ticker    *time.Ticker
	in        chan any
	out       chan any
	buffer    []any
	mu        sync.Mutex
	done      chan struct{}
}

// Verify Batch implements the streams.Flow interface
var _ streams.Flow = (*Batch[any])(nil)

// NewBatch returns a new Batch instance
//
// it will batch the incoming elements into a slice
// whether the batch size is reached or the batch time is elapsed.
func NewBatch[T any](batchSize uint, batchTime time.Duration) *Batch[T] {
	if batchSize <= 0 {
		panic("batch size should greater than 0")
	}

	bf := &Batch[T]{
		batchSize: batchSize,
		batchTime: batchTime,
		ticker:    time.NewTicker(batchTime),
		buffer:    make([]any, 0, batchSize),
		in:        make(chan any),
		out:       make(chan any),
		done:      make(chan struct{}),
	}

	go bf.doStream()

	return bf
}

// Via sends the flow to the next stage via specified flow
func (bf *Batch[T]) Via(flow streams.Flow) streams.Flow {
	go bf.transmit(flow)

	return flow
}

// To sends the flow to the given sink
func (bf *Batch[T]) To(sink streams.Sink) {
	bf.transmit(sink)
}

// Out returns the output channel of the Batch
func (bf *Batch[T]) Out() <-chan any {
	return bf.out
}

// In returns the input channel of the Batch
func (bf *Batch[T]) In() chan<- any {
	return bf.in
}

func (bf *Batch[T]) doStream() {
	batches := make([]T, 0, bf.batchSize)
	tick := time.NewTicker(bf.batchTime)

	defer func() {
		tick.Stop()
		if len(batches) > 0 {
			bf.out <- batches
		}
		close(bf.out)
	}()

	for {
		select {
		case v, ok := <-bf.in:
			if ok {
				batches = append(batches, v.(T))
				if len(batches) >= int(bf.batchSize) {
					bf.out <- batches
					tick.Reset(bf.batchTime)
					batches = make([]T, 0, bf.batchSize)
				}
			} else {
				return
			}
		case <-tick.C:
			if len(batches) > 0 {
				bf.out <- batches
				batches = make([]T, 0, bf.batchSize)
			}
		}
	}
}

func (bf *Batch[T]) transmit(inlet streams.Inlet) {
	for item := range bf.out {
		inlet.In() <- item
	}
	defer close(inlet.In())
}
