package processor

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/parnurzeal/gorequest"

	"vec/db"
	"vec/model"
)

type strToVec struct {
	reqUrl    string
	field     string
	vecWriter io.ReadWriteCloser
}

// ToVec 这里不能用指针 receiver，否则后面循环的时候，可能会导致 p.vecWriter 一直是最后一个文件的指针
func (p strToVec) ToVec(patent *model.Patent) vector {
	request := gorequest.New()

	res := &struct {
		Data []string `json:"data"`
	}{}
	payload := map[string]interface{}{
		"strarr": []string{patent.GetField(p.field)},
	}
	resp, body, errs := request.Post(p.reqUrl).
		Send(payload).
		EndStruct(res)
	if len(errs) > 0 {
		fmt.Println("errs: ", errs)
		panic(errs)
	}
	if resp.StatusCode != http.StatusOK {
		panic("status code != 200, body: " + string(body))
	}
	if len(res.Data) < 1 {
		fmt.Println(string(body))
		panic("response data length < 1")
	}
	return vector(fmt.Sprintf("%s-%s-%s-%s-vector", p.reqUrl, patent.ID, p.field, res.Data[0]))
}

func (p strToVec) SaveVec(vec vector) error {
	_, err := p.vecWriter.Write([]byte(vec.string() + "\n"))
	return err
}

func (p strToVec) SaveVecIDAndPatentID(patent *model.Patent, vecID int) error {
	return model.SetRedis(db.Get().Redis, p.addPrefixToVecID(vecID), patent.ID)
}

func (p strToVec) addPrefixToVecID(vecID int) string {
	return fmt.Sprintf("%s-%d", p.field, vecID)
}

// NewStrToVec 传入字段名和接口地址，返回一个处理器
func NewStrToVec(filed, reqUrl string) Processor {
	var vecFilename = fmt.Sprintf("%s.vec", filed)
	vecFile, err := os.OpenFile(vecFilename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	return &strToVec{
		field:     filed,
		reqUrl:    reqUrl,
		vecWriter: vecFile,
	}
}
