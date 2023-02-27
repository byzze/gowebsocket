/**
* Created by GoLand.
* User: link1st
* Date: 2019-07-25
* Time: 09:59
 */

package main

import (
	"fmt"
	"gowebsocket/lib/redislib"
	"gowebsocket/routers"
	"gowebsocket/servers/grpcserver"
	"gowebsocket/servers/task"
	"gowebsocket/servers/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 初始化配置文件
	initConfig()
	// 初始化日志文件
	initFile()
	// 初始化redis
	initRedis()

	router := gin.Default()
	// 初始化路由
	routers.Init(router)
	routers.WebsocketInit()

	// 定时任务
	task.Init()

	// 服务注册
	task.ServerInit()
	// websocket
	go websocket.StartWebSocket()
	// grpc
	go grpcserver.Init()

	go open()
	// http
	httpPort := viper.GetString("app.httpPort")
	http.ListenAndServe(":"+httpPort, router)

}

// 初始化日志
func initFile() {
	// Disable Console Color, you don't need console color when writing the logs to file.
	gin.DisableConsoleColor()

	// 获取日志文件配置路径
	logFile := viper.GetString("app.logFile")
	f, err := os.Create(logFile)
	if err != nil {
		log.Println(err)
	}
	gin.DefaultWriter = io.MultiWriter(f)
}

func initConfig() {
	// 读取配置文件
	viper.SetConfigName("config/app")
	viper.AddConfigPath(".") // 添加搜索路径

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	// 获取app属性，redis属性
	fmt.Println("config app:", viper.Get("app"))
	fmt.Println("config redis:", viper.Get("redis"))

}

func initRedis() {
	redislib.ExampleNewClient()
}

func open() {

	time.Sleep(1000 * time.Millisecond)

	httpUrl := viper.GetString("app.httpUrl")
	httpUrl = "http://" + httpUrl + "/home/index"

	fmt.Println("访问页面体验:", httpUrl)
	// 程序启动自动打开网址
	cmd := exec.Command("open", httpUrl)
	cmd.Output()
}
