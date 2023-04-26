// Copyright (c) 2021 Cisco All Rights Reserved.

package conf

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init() {

	viper.AddConfigPath("./conf/")
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Error reading config file, %s", err)
	}

	viper.SetDefault("logging.level", "info")

	level, lErr := log.ParseLevel(viper.GetString("logging.level"))
	if lErr != nil {
		panic("unrecognized loglevel")
	}
	log.SetLevel(level)

	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())
	viper.SetDefault("message.count", 0)
	viper.SetDefault("message.randomise", true)
	viper.SetDefault("message.event-per-sec", 2)
	viper.SetDefault("message.byte-per-sec", 200)

	viper.SetDefault("api.addr", ":11000")
	viper.SetDefault("api.basePath", "/")

	viper.SetDefault("nginx.enabled", false)
	viper.SetDefault("nginx.output_format", "{{.Remote}} {{.Host}} {{.User}} [{{.Time}}] \"{{.Method}} {{.Path}} HTTP/1.1\" {{.Code}} {{.Size}} \"{{.Referer}}\" \"{{.Agent}}\" \"{{.HttpXForwardedFor}}\"")
	viper.SetDefault("nginx.time_format", "02/Jan/2006:15:04:05 -0700")

	viper.SetDefault("apache.enabled", false)
	viper.SetDefault("apache.output_format", "{{.Remote}} {{.Host}} {{.User}} [{{.Time}}] {{.User}} \"{{.Method}} {{.Path}} HTTP/1.1\" {{.Code}} {{.Size}} \"{{.Referer}}\" \"{{.Agent}}\" \"{{.HttpXForwardedFor}}\"")
	viper.SetDefault("apache.time_format", "02/Jan/2006:15:04:05 -0700")

	viper.SetDefault("golang.enabled", false)
	viper.SetDefault("golang.output_format", "{{.Environment}} {{.Application}} {{.Component}} [{{.Time}}] {{.Level}} \"{{.MSG}}\"")
	viper.SetDefault("golang.time_format", "02/Jan/2006:15:04:05 -0700")
	viper.SetDefault("golang.weight.error", 0)
	viper.SetDefault("golang.weight.info", 1)
	viper.SetDefault("golang.weight.warning", 0)
	viper.SetDefault("golang.weight.debug", 0)
}
