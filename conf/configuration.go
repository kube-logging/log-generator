// Copyright © 2021 Cisco Systems, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License."""

package conf

import (
	"fmt"

	"github.com/go-viper/encoding/ini"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Viper *viper.Viper

func Init() {
	codecRegistry := viper.NewCodecRegistry()
	codecRegistry.RegisterCodec("ini", &ini.Codec{})
	Viper = viper.NewWithOptions(
		viper.WithCodecRegistry(codecRegistry),
	)
	Viper.SetConfigName("config")
	Viper.SetConfigType("ini")
	Viper.AddConfigPath("./conf/")

	if err := Viper.ReadInConfig(); err != nil {
		log.Warnf("Error reading config file, %s", err)
	}

	Viper.SetDefault("logging.level", "info")

	level, lErr := log.ParseLevel(Viper.GetString("logging.level"))
	if lErr != nil {
		panic("unrecognized loglevel")
	}
	log.SetLevel(level)

	fmt.Printf("Using config: %s\n", Viper.ConfigFileUsed())
	Viper.SetDefault("message.count", 0)
	Viper.SetDefault("message.randomise", true)
	Viper.SetDefault("message.event-per-sec", 2)
	Viper.SetDefault("message.byte-per-sec", 200)
	Viper.SetDefault("message.max-random-hosts", 1000)
	Viper.SetDefault("message.max-random-apps", 100)
	Viper.SetDefault("message.host", "hostname")
	Viper.SetDefault("message.appname", "appname")

	Viper.SetDefault("api.addr", ":11000")
	Viper.SetDefault("api.basePath", "/")

	Viper.SetDefault("nginx.enabled", false)
	Viper.SetDefault("apache.enabled", false)
	Viper.SetDefault("golang.enabled", false)
	Viper.SetDefault("golang.time_format", "02/Jan/2006:15:04:05 -0700")
	Viper.SetDefault("golang.weight.error", 0)
	Viper.SetDefault("golang.weight.info", 1)
	Viper.SetDefault("golang.weight.warning", 0)
	Viper.SetDefault("golang.weight.debug", 0)
}
