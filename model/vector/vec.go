package vector

type Vector struct {
	description string
	vectors     []float32
}

func NewVector(vectors []float32, description string) *Vector {
	return &Vector{
		description: description,
		vectors:     vectors,
	}
}

func (v *Vector) Describe() string {
	return v.description
}

func (v *Vector) Dim() int {
	return len(v.vectors)
}

func (v *Vector) Vectors() []float32 {
	return v.vectors
}
