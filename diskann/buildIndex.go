package diskann

import (
	"fmt"

	"github.com/parnurzeal/gorequest"

	"vec/config"
	"vec/model/vector"
)

// BuildIndex 调用 diskann 上游服务，为所有字段建立索引
func BuildIndex() {
	for _, field := range config.Get().Str2VecConfigs {
		BuildOneIndex(field.Field)
	}
}

// BuildOneIndex 调用 diskann 上游服务，为特定的字段建立索引
func BuildOneIndex(field string) {
	buildOneIndex(field, vector.GetIndexVectorFullPath(field), config.Get().DiskannConfig.BuildIndexUrl)
}

// 调用 diskann 上游服务，使用 vectorFile 建立索引
// 并且指定索引的名称为 indexName，后面查询的时候使用相同的 indexName
func buildOneIndex(indexName, vectorFile, buildIndexUrl string) {
	fmt.Printf("已经请求了 diskann 的 buildIndex 服务，url: %s, fvec: %s, field: %s\n", buildIndexUrl, vectorFile, indexName)
	// 因为响应是异步的，所以这里不需要等待响应
	gorequest.New().Post(buildIndexUrl).Send(map[string]interface{}{
		"fvec":  vectorFile,
		"field": indexName,
	}).End()
}
