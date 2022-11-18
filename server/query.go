package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"

	"vec/config"
	"vec/model"
	"vec/processor"
)

func query(c *gin.Context) {
	// 处理请求参数
	req := struct {
		Field string `json:"field"`
		Query string `json:"query"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil || req.Field == "" || req.Query == "" {
		Fail(c, 40001, "filed or query is empty")
		return
	}

	// 根据字段和查询内容拼接专利内容
	patent := model.FiledToPatent(req.Field, req.Query)
	if patent == nil {
		Fail(c, 40002, "invalid field")
		return
	}
	var process processor.Processor
	// 使用正则去除 req.Query 中的所有空白字符
	queryWithoutBlank := regexp.MustCompile(`\s+`).ReplaceAllString(req.Query, "")
	// 拼接查询向量文件的路径
	vecFileName := fmt.Sprintf("%s.vec", queryWithoutBlank)
	path := filepath.Join(config.Get().ServerConfig.VectorDir, queryVecDir, vecFileName)
	// 根据不同的字段，使用不同的处理器
	for _, conf := range config.Get().Str2VecConfigs {
		if conf.Field == req.Field {
			file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				Fail(c, 50001, "open file error: "+err.Error())
				return
			}
			process = processor.NewStrToVecWithVecWriter(conf.Field, conf.Url, file)
			break
		}
	}
	if process == nil {
		Fail(c, 40003, "field is not defined in config")
		return
	}
	// 生成查询向量
	vec := process.ToVec(patent)
	// 保存查询向量到指定文件
	err := process.SaveVec(vec)
	if err != nil {
		Fail(c, 50002, "save vec error")
		return
	}

	vectorIDs, err := queryDiskann(path)
	if err != nil {
		Fail(c, 50003, err.Error())
		return
	}

	// 根据向量 id 查询专利
	patentIDs := make([]string, 0, len(vectorIDs))
	for _, vectorID := range vectorIDs {
		patentID, err := model.GetPatentIDByVectorID(req.Field, vectorID)
		if err != nil {
			Fail(c, 50004, fmt.Errorf("get patent id by vector id error: %v", err).Error())
			return
		}
		patentIDs = append(patentIDs, patentID)
	}
	Success(c, patentIDs)
}

// 根据 fevs 文件，调用 diskann，获得最相似的专利向量的 id
func queryDiskann(vecFile string) ([]string, error) {
	url := "http://10.101.32.33:18180/SearchDiskIndex"
	data := map[string]any{
		"fvec": vecFile,
	}
	res := struct {
		Data []string `json:"data"`
	}{}
	httpResp, body, errs := gorequest.New().Post(url).Send(data).EndStruct(&res)
	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code is %d", httpResp.StatusCode)
	}
	if len(errs) > 0 {
		fmt.Printf("query diskann struct parse error, response body: %v\n", body)
		return nil, fmt.Errorf("query diskann error: %v", errs)
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf("no vector id found")
	}
	fmt.Println(res.Data)
	return res.Data, nil
}
