package conf

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init() {

	viper.AddConfigPath("./conf/")
	viper.SetConfigName("config")
	viper.ReadInConfig()

	//log.SetFormatter(&log.TextFormatter{})
	//fmt.Printf(">>>>>%s ",viper.GetString("logging.level"))
	viper.SetDefault("logging.level", "info")
	switch viper.GetString("logging.level") {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		panic("unrecognized loglevel")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	fmt.Printf("Using config: %s\n", viper.ConfigFileUsed())
	viper.SetDefault("message.count", 10)
	viper.SetDefault("message.randomise", true)
	viper.SetDefault("message.event-per-sec", 2)
	viper.SetDefault("message.byte-per-sec", 200)

	viper.SetDefault("metrics.addr", ":11000")
	viper.SetDefault("metrics.path", "/metrics")

	viper.SetDefault("nginx.enabled", true)
	viper.SetDefault("apache.enabled", true)


}
