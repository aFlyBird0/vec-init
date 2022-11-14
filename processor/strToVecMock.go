package processor

import (
	"fmt"
	"io"
	"os"

	"vec/model"
)

type strToVecMock struct {
	reqUrl    string
	field     string
	vecWriter io.ReadWriteCloser
}

// ToVec 这里不能用指针 receiver，否则后面循环的时候，可能会导致 p.vecWriter 一直是最后一个文件的指针
func (p strToVecMock) ToVec(patent *model.Patent) vector {
	return vector(fmt.Sprintf("%s-%s-%s-%s-vector", p.reqUrl, patent.ID, p.field, patent.GetField(p.field)))
}

func (p strToVecMock) SaveVec(vec vector) error {
	_, err := p.vecWriter.Write([]byte(vec.string() + "\n"))
	return err
}

func (p strToVecMock) SaveVecIDAndPatentID(patent *model.Patent, vecID int) error {
	//return model.SetRedis(db.Get().Redis, p.addPrefixToVecID(vecID), patent.ID)
	return nil
}

func (p strToVecMock) addPrefixToVecID(vecID int) string {
	return fmt.Sprintf("%s-%d", p.field, vecID)
}

// NewStrToVec 传入字段名和接口地址，返回一个处理器
func NewStrToVecMock(filed, reqUrl string) Processor {
	var vecFilename = fmt.Sprintf("%s.vec", filed)
	vecFile, err := os.OpenFile(vecFilename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	return &strToVecMock{
		field:     filed,
		reqUrl:    reqUrl,
		vecWriter: vecFile,
	}
}

func NewNameToVecMock() Processor {
	return NewStrToVecMock("name", "http://name.to.vec")
}

func NewAbstractToVecMock() Processor {
	return NewStrToVecMock("abstract", "http://abstract.to.vec")
}
