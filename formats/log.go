// Copyright (c) 2023 Cisco All Rights Reserved.

package formats

import (
	"io/fs"
	"log"
	"strings"
	"text/template"

	"github.com/kube-logging/log-generator/formats/golang"
	"github.com/kube-logging/log-generator/formats/syslog"
	"github.com/kube-logging/log-generator/formats/web"
	"github.com/prometheus/client_golang/prometheus"
)

// TODO: factory
type void = struct{}

var Types = map[string]struct{}{
	"golang": void{},
	"syslog": void{},
	"web":    void{},
}

type Log interface {
	String() (string, float64)
	Labels() prometheus.Labels
}

type LogTemplateData interface {
	Severity() string
}

type LogTemplate struct {
	Format string

	template *template.Template
	data     LogTemplateData
}

func NewSyslog(format string) (*LogTemplate, error) {
	return newLogTemplate(format, syslog.TemplateFS, syslog.SampleData())
}

func NewRandomSyslog(format string) (*LogTemplate, error) {
	return newLogTemplate(format, syslog.TemplateFS, syslog.RandomData())
}

func SyslogFormatNames() []string {
	return formatNames(syslog.TemplateFS)
}

func NewWeb(format string) (*LogTemplate, error) {
	return newLogTemplate(format, web.TemplateFS, web.SampleData())
}

func NewRandomWeb(format string) (*LogTemplate, error) {
	return newLogTemplate(format, web.TemplateFS, web.RandomData())
}

func WebFormatNames() []string {
	return formatNames(web.TemplateFS)
}

func NewGolangRandom(i golang.GolangLogIntensity) Log {
	return golang.NewGolangLogRandom(i)
}

func newLogTemplate(format string, fs fs.FS, data LogTemplateData) (*LogTemplate, error) {
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

func formatNames(fs fs.FS) []string {
	formats := []string{}

	for _, t := range loadAllTemplates(fs) {
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
