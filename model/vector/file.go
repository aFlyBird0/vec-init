package vector

import (
	"os"
	"path/filepath"

	"vec/config"
)

const (
	FileExt = ".fvecs"

	initVecSubDir = "init"
	queryVecDir   = "query"
)

// AddVectorExt 为文件路径添加向量文件后缀
func AddVectorExt(path string) string {
	return path + FileExt
}

// GetIndexVectorFullPath 拼接/获取 初始化(索引)向量文件的路径
func GetIndexVectorFullPath(field string) string {
	pathWithoutExt := filepath.Join(config.Get().ServerConfig.VectorDir, initVecSubDir, field)
	return AddVectorExt(pathWithoutExt)
}

// GetQueryVectorFullPath 拼接/获取 查询向量文件的路径
func GetQueryVectorFullPath(query string) string {
	pathWithoutExt := filepath.Join(config.Get().ServerConfig.VectorDir, queryVecDir, query)
	return AddVectorExt(pathWithoutExt)
}

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
	err = os.MkdirAll(filepath.Join(absPath, queryVecDir), os.ModePerm)
	if err != nil {
		panic(err)
	}
}
