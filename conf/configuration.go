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
	viper.SetDefault("message.max-random-hosts", 1000)
	viper.SetDefault("message.max-random-apps", 100)
	viper.SetDefault("message.max-random-cap", 10000)

	viper.SetDefault("api.addr", ":11000")
	viper.SetDefault("api.basePath", "/")

	viper.SetDefault("nginx.enabled", false)
	viper.SetDefault("apache.enabled", false)
	viper.SetDefault("golang.enabled", false)
	viper.SetDefault("golang.time_format", "02/Jan/2006:15:04:05 -0700")
	viper.SetDefault("golang.weight.error", 0)
	viper.SetDefault("golang.weight.info", 1)
	viper.SetDefault("golang.weight.warning", 0)
	viper.SetDefault("golang.weight.debug", 0)
}
