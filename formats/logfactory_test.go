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
	"strings"
	"testing"
	"text/template"

	"github.com/kube-logging/log-generator/formats/syslog"
	"github.com/kube-logging/log-generator/formats/web"
	"github.com/kube-logging/log-generator/log"
)

type LogConstructor func(string) (*log.LogTemplate, error)

func TestSyslogFormats(t *testing.T) {
	// Separate Template tree per directory because Go Template can not handle
	// multiple files with the same name in different directories.

	syslogTemplates := log.LoadAllTemplates(syslog.TemplateFS)
	assertFormatAll(t, syslogTemplates, NewSyslog)
}

func TestWebFormats(t *testing.T) {
	// Separate Template tree per directory because Go Template can not handle
	// multiple files with the same name in different directories.

	webTemplates := log.LoadAllTemplates(web.TemplateFS)
	assertFormatAll(t, webTemplates, NewWeb)
}

func assertFormatAll(t *testing.T, templates []*template.Template, c LogConstructor) {
	for _, f := range templates {
		format := strings.TrimSuffix(f.Name(), ".tmpl")
		log, err := c(format)
		if err != nil {
			t.Fatalf("Failed to create log, format=%q, %v", format, err)
		}

		if l, _ := log.String(); len(strings.TrimSpace(l)) == 0 {
			t.Logf("Rendered log is empty, format=%q", format)
		}
	}
}
