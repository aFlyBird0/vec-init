package stream

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
)

type BinaryFileSink struct {
	fileName  string
	vecWriter io.Writer
	in        chan interface{}
}

// NewBinaryFileSink returns a new BinaryFileSink instance.
func NewBinaryFileSink(fileName string) *BinaryFileSink {
	sink := &BinaryFileSink{
		fileName: fileName,
		in:       make(chan interface{}),
	}
	sink.init()
	return sink
}

func (bfs *BinaryFileSink) init() {
	go func() {
		file, err := os.OpenFile(bfs.fileName, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("BinaryFileSink failed to open the file %s", bfs.fileName)
		}
		defer file.Close()
		bfs.vecWriter = file
		for elem := range bfs.in {
			err := bfs.saveVec(elem.(*VectorPatentAndVectorID))
			if err != nil {
				fmt.Printf("BinaryFileSink failed to write to the file %s", bfs.fileName)
			}
		}
	}()
}

// In returns an input channel for receiving data
func (bfs *BinaryFileSink) In() chan<- interface{} {
	return bfs.in
}

func (bfs *BinaryFileSink) saveVec(vec *VectorPatentAndVectorID) error {
	// fvecs 文件格式，对于每个向量：
	// 1. 先写入4字节的整数dim，表示向量的维度
	// 2. 再依次写入dim*4字节的浮点数，即向量的每个维度的值
	// 再写入下一个向量，向量各维度之间、向量间无分隔符
	err1 := binary.Write(bfs.vecWriter, binary.LittleEndian, int32(vec.Dim()))
	err2 := binary.Write(bfs.vecWriter, binary.LittleEndian, vec.Vectors())
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
