package main

import (
	"fmt"
	"os"
	"xiaodainiao/config"
	"xiaodainiao/lib/logger"
	"xiaodainiao/resp/handler"
	"xiaodainiao/tcp"
)

const configFile string = "redis.conf"

//如果没有配置文件，默认就是下面这个
var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

func fileExits(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	//首先设置日志的格式
	logger.Info(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2020-01-02",
	})
	if fileExits(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	err := tcp.ListenAndServerWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
		},
		handler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
