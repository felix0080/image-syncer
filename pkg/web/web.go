package web

import (
	"fmt"
	//"github.com/DeanThompson/ginpprof"
	"image-syncer/pkg/client"
	"image-syncer/pkg/object"
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/DeanThompson/ginpprof"
	"time"
)

func HttpServer() *http.Server {
	router := gin.Default()
	monitor := router.Group("/image")
	monitor.POST("/download", download)
	monitor.POST("/find", find)
	ginpprof.Wrap(router)
	fmt.Println("start ")
	server := &http.Server{
		Addr:         ":8888",
		WriteTimeout: time.Minute * 120,
		IdleTimeout: time.Minute * 120,
		ReadTimeout: time.Minute * 120,
		Handler:      router,
	}
	log.Println("server prepare")
	return server
}
/**
list 前端显示调用
*/
func HttpStart(s *http.Server) (stop chan struct{}) {
	c := make(chan struct{}, 1)
	go func(ch chan struct{}) {
		err := s.ListenAndServe()
		if err != nil {
			fmt.Println(err)
			ch <- struct{}{}
			return
		}
		ch <- struct{}{}
	}(c)
	return c
}
//同步下载镜像，同步返回下载结果
func download(c *gin.Context) {
	before:=time.Now()
	var rb object.Param
	if err := c.ShouldBindJSON(&rb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rb.SelectSysCode()
	task,err:=client.Clients.GenerateSyncTaskSync(rb)
	du:=time.Now().Sub(before)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"result":"failure","duration":du.Seconds()})
		return
	}
	err=task.Run()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"result":"failure","duration":du.Seconds()})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"error":nil,
		"result":"success",
		"duration":du.Seconds(),
	})
}
//同步下载镜像，同步返回下载结果
func find(c *gin.Context) {
	before:=time.Now()
	var rb object.Param
	if err := c.ShouldBindJSON(&rb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"result":false})
		return
	}
	rb.SelectSysCode()
	b,err:=client.Clients.CheckIfExist(rb)
	du:=time.Now().Sub(before)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"result":b,"duration":du.Seconds()})
		return
	}
	c.JSON(http.StatusOK,gin.H{"error": nil,"result":b,"duration":du.Seconds()})
}