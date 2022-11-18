package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"vec/config"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

const queryVecDir = "query"

func Serve() {
	Init()
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/query", query)

	host := config.Get().ServerConfig.Host
	port := config.Get().ServerConfig.Port

	err := r.Run(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "ok",
		Data: data,
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(code/100, Response{
		Code: code,
		Msg:  msg,
	})
}

func Init() {
	// 初始化向量文件夹
	absPath, err := filepath.Abs(config.Get().VectorDir)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(filepath.Join(absPath, queryVecDir), os.ModePerm)
	if err != nil {
		panic(err)
	}
}
