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

package golang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Pallinder/go-randomdata"
	wr "github.com/mroth/weightedrand"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/kube-logging/log-generator/conf"
)

type GolangLogIntensity struct {
	ErrorWeight   *uint `json:"error_weight"`
	WarningWeight *uint `json:"warning_weight"`
	InfoWeight    *uint `json:"info_weight"`
	DebugWeight   *uint `json:"debug_weight"`
}

type GolangLog struct {
	Application string `json:"application"`
	Environment string `json:"environment"`
	Component   string `json:"component"`
	Level       string `json:"level"`
	MSG         string `json:"msg"`
	Time        string `json:"time"`

	isFramed bool
}

func (g GolangLog) newRandomMessage() string {
	msgList := map[string]string{
		"info": randomdata.StringSample(
			"constructing many client instances from the same exec auth config can cause performance problems during cert rotation and can exhaust available network connections; 1083 clients constructed calling",
			"starting posthook function",
			"rbac.authorization.k8s.io/v1beta1 ClusterRole is deprecated in v1.17+, unavailable in v1.22+; use rbac.authorization.k8s.io/v1 ClusterRole\\n",
		),
		"warning": randomdata.StringSample(
			"no security scan whitelist information available...",
			"firewall is still alive",
			"apiextensions.k8s.io/v1beta1 CustomResourceDefinition is deprecated in v1.16+, unavailable in v1.22+; use apiextensions.k8s.io/v1 CustomResourceDefinition",
		),
		"error": randomdata.StringSample(
			"could not get cluster from database: could not find cluster by ID: cluster not found",
			"Activity error.",
			"converting cluster model to common cluster failed: record not found",
		),
	}
	return msgList[g.Level]
}

func NewGolangLogRandom(i GolangLogIntensity) *GolangLog {
	rand.Seed(time.Now().UTC().UnixNano())
	c, err := wr.NewChooser(
		wr.Choice{Item: "error", Weight: *i.ErrorWeight},
		wr.Choice{Item: "warning", Weight: *i.WarningWeight},
		wr.Choice{Item: "info", Weight: *i.InfoWeight},
		wr.Choice{Item: "debug", Weight: *i.DebugWeight},
	)
	if err != nil {
		log.Error(err)
	}
	return &GolangLog{
		Application: randomdata.StringSample("webshop", "blog"),
		Environment: randomdata.StringSample("production", "sandbox", "demo"),
		Component:   randomdata.StringSample("frontend", "backend", "worker"),
		Level:       c.Pick().(string),
		Time:        "",
	}
}

func (g GolangLog) String() (string, float64) {
	g.Time = time.Now().Format(conf.Viper.GetString("golang.time_format"))
	g.MSG = g.newRandomMessage()

	out, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		log.Error(err)
	}

	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, out); err != nil {
		log.Error(err)
	}

	message := fmt.Sprint(buffer.String())

	return message, float64(len([]byte(message)))
}

func (l *GolangLog) IsFramed() bool {
	return l.isFramed
}

func (l *GolangLog) SetFramed(f bool) {
	l.isFramed = f
}

func (g GolangLog) Labels() prometheus.Labels {
	return prometheus.Labels{
		"type":     "golang",
		"severity": g.Level,
	}
}
