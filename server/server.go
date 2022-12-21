package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"vec/config"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Serve() {
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
