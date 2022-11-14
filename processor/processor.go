package processor

import (
	"vec/model"
)

//type Processor interface {
//	Process(p *model.Patent, vecID int) error
//}

type Processor interface {
	ToVec(p *model.Patent) vector
	SaveVec(vec vector) error
	SaveVecIDAndPatentID(p *model.Patent, vecID int) error
}
