package processor

import (
	"os"
	"path/filepath"

	"vec/config"
)

func Init() {
	// 初始化向量文件夹
	absPath, err := filepath.Abs(config.Get().VectorDir)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(filepath.Join(absPath, initVecSubDir), os.ModePerm)
	if err != nil {
		panic(err)
	}
}
