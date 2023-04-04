package main

import (
	"fmt"
	"os"

	"github.com/LynchQ/my-go-redis/config"
	"github.com/LynchQ/my-go-redis/lib/logger"
	"github.com/LynchQ/my-go-redis/tcp"
	EchoHandler "github.com/LynchQ/my-go-redis/tcp"
)

const configFile string = "redis.conf"

// 默认配置
var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	// q: && !info.IsDir()是什么意思？
	// a: 如果err == nil && !info.IsDir()为真，说明文件存在且不是目录
	return !info.IsDir() && err == nil
}
func main() {
	// 初始化日志
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "mygoredis",
		Ext:        "log",
		TimeFormat: "1996-09-28",
	})

	// 如果配置文件存在，则读取配置文件
	if fileExists(configFile) {
		config.SetupConfig(configFile)
		// 如果配置文件不存在，则使用默认配置
	} else {
		config.Properties = defaultProperties
	}

	// 启动服务
	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d",
				config.Properties.Bind,
				config.Properties.Port),
		},
		EchoHandler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}

}
