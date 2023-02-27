package server

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/parnurzeal/gorequest"

	"vec/config"
	"vec/model"
	"vec/model/vector"
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
	vecFilePath := vector.GetQueryVectorFullPath(queryWithoutBlank)
	// 根据不同的字段，使用不同的处理器
	for _, conf := range config.Get().Str2VecConfigs {
		if conf.Field == req.Field {
			file, err := os.OpenFile(vecFilePath, os.O_CREATE|os.O_WRONLY, 0666)
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

	vectorIDScores, err := queryDiskann(vecFilePath, req.Field)
	if err != nil {
		Fail(c, 50003, err.Error())
		return
	}

	// 根据向量 id 查询专利
	patentIDScores := make([]IDScore, 0, len(vectorIDScores))
	for _, idScore := range vectorIDScores {
		id, err := strconv.ParseInt(idScore.ID, 10, 64)
		if err != nil {
			fmt.Printf("err parsing ids from diskann response, response: %v, err: %v\n", patentIDScores, err)
		}
		patentID, err := model.GetPatentIDByVectorID(req.Field, id)
		if err != nil {
			Fail(c, 50004, fmt.Errorf("get patent id by vector id error: %v", err).Error())
			return
		}
		if patentID != "" {
			patentIDScores = append(patentIDScores, IDScore{
				ID:    patentID,
				Score: idScore.Score,
			})
		} else {
			fmt.Printf("vector id <%s-%s> not found in db\n", req.Field, idScore.ID)
		}
	}
	Success(c, patentIDScores)
}

type IDScore struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}

// 根据 fevs 文件，指定索引名（field），调用 diskann，获得最相似的专利向量的 id 和相似度
func queryDiskann(vecFile, field string) ([]IDScore, error) {
	url := config.Get().DiskannConfig.QueryUrl
	data := map[string]any{
		"fvec":  vecFile,
		"field": field,
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
	// 返回的数据是一个 string 数组，[id1, score1, id2, score2, ...]
	if len(res.Data)%2 != 0 {
		return nil, fmt.Errorf("id and score count not match")
	}

	idScores := make([]IDScore, 0, len(res.Data)/2)

	for i := 0; i < len(res.Data); i += 2 {
		score, err := strconv.ParseFloat(res.Data[i+1], 64)
		if err != nil {
			return nil, fmt.Errorf("parse score error, data: %v, id: %s, score: %s, err: %v", res.Data, res.Data[i], res.Data[i+1], err)
		}
		idScores = append(idScores, IDScore{
			ID:    res.Data[i],
			Score: score,
		})
	}
	return idScores, nil
}
