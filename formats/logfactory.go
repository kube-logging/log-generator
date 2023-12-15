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
	"embed"

	"github.com/kube-logging/log-generator/formats/custom"
	"github.com/kube-logging/log-generator/formats/golang"
	"github.com/kube-logging/log-generator/formats/web"
	"github.com/kube-logging/log-generator/log"
)

func FormatsByType() map[string][]string {
	response := map[string][]string{}
	for t, f := range custom.Formats() {
		response[t] = f
	}
	response["web"] = WebFormatNames()
	return response
}

func LogFactory(logType string, format string, randomise bool) (log.Log, error) {
	switch logType {
	case "web":
		if randomise {
			return NewRandomWeb(format, web.TemplateFS)
		} else {
			return NewWeb(format, web.TemplateFS)
		}
	default:
		return custom.LogFactory(logType, format, randomise)
	}
}

func NewWeb(format string, templates embed.FS) (*log.LogTemplate, error) {
	return log.NewLogTemplate(format, templates, web.SampleData())
}

func NewRandomWeb(format string, templates embed.FS) (*log.LogTemplate, error) {
	return log.NewLogTemplate(format, templates, web.RandomData())
}

func WebFormatNames() []string {
	return log.FormatNames(web.TemplateFS)
}

func NewGolangRandom(i golang.GolangLogIntensity) log.Log {
	return golang.NewGolangLogRandom(i)
}
