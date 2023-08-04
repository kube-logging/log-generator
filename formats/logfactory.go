// Copyright © 2023 Kube logging authors
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

package formats

import (
	"github.com/kube-logging/log-generator/formats/custom"
	"github.com/kube-logging/log-generator/formats/golang"
	"github.com/kube-logging/log-generator/formats/syslog"
	"github.com/kube-logging/log-generator/formats/web"
	"github.com/kube-logging/log-generator/log"
)

func FormatsByType() map[string][]string {
	response := map[string][]string{}
	for t, f := range custom.Formats() {
		response[t] = f
	}
	response["syslog"] = SyslogFormatNames()
	response["web"] = WebFormatNames()
	return response
}

func LogFactory(logType string, format string, randomise bool) (log.Log, error) {
	switch logType {
	case "syslog":
		if randomise {
			return NewRandomSyslog(format)
		} else {
			return NewSyslog(format)
		}
	case "web":
		if randomise {
			return NewRandomWeb(format)
		} else {
			return NewWeb(format)
		}
	default:
		return custom.LogFactory(logType, format, randomise)
	}
}

func NewSyslog(format string) (*log.LogTemplate, error) {
	return log.NewLogTemplate(format, syslog.TemplateFS, syslog.SampleData())
}

func NewRandomSyslog(format string) (*log.LogTemplate, error) {
	return log.NewLogTemplate(format, syslog.TemplateFS, syslog.RandomData())
}

func SyslogFormatNames() []string {
	return log.FormatNames(syslog.TemplateFS)
}

func NewWeb(format string) (*log.LogTemplate, error) {
	return log.NewLogTemplate(format, web.TemplateFS, web.SampleData())
}

func NewRandomWeb(format string) (*log.LogTemplate, error) {
	return log.NewLogTemplate(format, web.TemplateFS, web.RandomData())
}

func WebFormatNames() []string {
	return log.FormatNames(web.TemplateFS)
}

func NewGolangRandom(i golang.GolangLogIntensity) log.Log {
	return golang.NewGolangLogRandom(i)
}
