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

package log

import (
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

type LogTemplateData interface {
	Severity() string
}

type LogTemplate struct {
	Format string

	template *template.Template
	data     LogTemplateData
}

func NewLogTemplate(format string, fs fs.FS, data LogTemplateData) (*LogTemplate, error) {
	template, err := loadTemplate(format, fs)
	if err != nil {
		return nil, err
	}

	return &LogTemplate{
		Format:   format,
		template: template,
		data:     data,
	}, nil
}

func FormatNames(fs fs.FS) []string {
	formats := []string{}

	for _, t := range LoadAllTemplates(fs) {
		formats = append(formats, strings.TrimSuffix(t.Name(), ".tmpl"))
	}

	return formats
}

func (l *LogTemplate) String() (string, float64) {
	var b strings.Builder
	if err := l.template.Execute(&b, l.data); err != nil {
		log.Panic(err.Error())
	}

	str := strings.TrimSuffix(b.String(), "\n")

	return str, float64(len([]byte(str)))
}

func (l *LogTemplate) Labels() prometheus.Labels {
	return prometheus.Labels{
		"type":     l.Format,
		"severity": l.data.Severity(),
	}
}

func loadTemplate(name string, fs fs.FS) (*template.Template, error) {
	// syslog.rfc5424.sdata => syslog.tmpl
	templateFileName, _, structuredFormatName := strings.Cut(name, ".")
	templateFileName += ".tmpl"

	template, err := template.ParseFS(fs, templateFileName)
	if err != nil {
		return nil, fmt.Errorf("could not parse format %q, %v", name, err)
	}

	if t := template.Lookup(name); t != nil {
		t.Option("missingkey=error")
		return t, nil
	}

	if t := template.Lookup(templateFileName); t != nil && !structuredFormatName {
		t.Option("missingkey=error")
		return t, nil
	}

	return nil, fmt.Errorf("could not find format %q", name)
}

func LoadAllTemplates(fs fs.FS) []*template.Template {
	return template.Must(template.ParseFS(fs, "*.tmpl")).Templates()
}
