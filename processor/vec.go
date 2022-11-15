package processor

// todo 重构至独立的包内
type vector struct {
	description string
	vectors     []float32
}

func NewVector(vectors []float32, description string) *vector {
	return &vector{
		description: description,
		vectors:     vectors,
	}
}

func (v *vector) describe() string {
	return v.description
}

func (v *vector) dim() int {
	return len(v.vectors)
}

func (v *vector) values() []float32 {
	return v.vectors
}
