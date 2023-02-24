package stream

import (
	"vec/model"
	"vec/model/vector"
)

type VectorPatent struct {
	*model.Patent
	*vector.Vector
}

type VectorPatentAndVectorID struct {
	*VectorPatent
	VectorID int64
}
