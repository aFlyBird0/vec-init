package processor

import (
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/parnurzeal/gorequest"

	"vec/db"
	"vec/model"
	"vec/model/vector"
)

type strToVec struct {
	reqUrl    string
	field     string
	vecWriter io.ReadWriteCloser
}

// ToVec 这里不能用指针 receiver，否则后面循环的时候，可能会导致 p.vecWriter 一直是最后一个文件的指针
func (p strToVec) ToVec(patent *model.Patent) *vector.Vector {
	request := gorequest.New()

	// todo 联系上游接口添加 code, msg 字段
	res := &struct {
		// 二维切片
		// 内层是向量，每个向量由n维的浮点数构成
		// 外层是向量的集合，每个查询的字符串都会返回一个对应的向量
		// 其实模型服务端传的是双精度的浮点数
		// 但是diskann只支持单精度的浮点数，所以这里舍弃了精度
		// 例如：[ [1.0, 2.0, 3.0], [4.0, 5.0, 6.0] ]
		// 一共有两个向量，每个向量有三个维度
		// 目前的生产环境是 1 个向量，768 维
		Data [][]float32 `json:"data"`
	}{}
	payload := map[string]interface{}{
		// 传入的字符串数组，目前先每次数组里只查询一个字符串
		"strarr": []string{patent.GetField(p.field)},
	}
	resp, body, errs := request.Post(p.reqUrl).
		Send(payload).
		EndStruct(res)
	if resp.StatusCode != http.StatusOK {
		panic("status code != 200, body: " + string(body))
	}
	if len(errs) > 0 {
		fmt.Println("errs: ", errs)
		panic(errs)
	}
	if len(res.Data) < 1 {
		fmt.Println(string(body))
		panic("response data length < 1")
	}
	for _, v := range res.Data {
		if len(v) < 1 {
			fmt.Println(string(body))
			panic("response data length < 1")
		}
	}

	// 目前每次只查询一个字符串，所以这里只有一个向量
	vector0 := res.Data[0]
	return vector.NewVector(vector0, fmt.Sprintf("%s-%s", p.field, patent.ID))
}

func (p strToVec) SaveVec(vec *vector.Vector) error {
	// fvecs 文件格式，对于每个向量：
	// 1. 先写入4字节的整数dim，表示向量的维度
	// 2. 再依次写入dim*4字节的浮点数，即向量的每个维度的值
	// 再写入下一个向量，向量各维度之间、向量间无分隔符
	err1 := binary.Write(p.vecWriter, binary.LittleEndian, int32(vec.Dim()))
	err2 := binary.Write(p.vecWriter, binary.LittleEndian, vec.Vectors())
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func (p strToVec) SaveVecIDAndPatentID(patent *model.Patent, vecID int) error {
	return model.SetRedis(db.Get().Redis, p.addPrefixToVecID(vecID), patent.ID)
}

func (p strToVec) addPrefixToVecID(vecID int) string {
	return fmt.Sprintf("%s-%d", p.field, vecID)
}

// NewStrToVec 传入字段名和接口地址，返回一个处理器
func NewStrToVec(field, reqUrl string) Processor {
	vecFilePath := vector.GetIndexVectorFullPath(field)
	vecFile, err := os.OpenFile(vecFilePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	return &strToVec{
		field:     field,
		reqUrl:    reqUrl,
		vecWriter: vecFile,
	}
}

func NewStrToVecWithVecWriter(filed, reqUrl string, vecWriter io.ReadWriteCloser) Processor {
	return &strToVec{
		field:     filed,
		reqUrl:    reqUrl,
		vecWriter: vecWriter,
	}
}
