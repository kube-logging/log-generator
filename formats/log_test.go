// Copyright (c) 2023 Cisco All Rights Reserved.

package formats

import (
	"strings"
	"testing"
	"text/template"

	"github.com/kube-logging/log-generator/formats/syslog"
	"github.com/kube-logging/log-generator/formats/web"
)

type LogConstructor func(string) (*LogTemplate, error)

func TestSyslogFormats(t *testing.T) {
	// Separate Template tree per directory because Go Template can not handle
	// multiple files with the same name in different directories.

	syslogTemplates := loadAllTemplates(syslog.TemplateFS)
	assertFormatAll(t, syslogTemplates, NewSyslog)
}

func TestWebFormats(t *testing.T) {
	// Separate Template tree per directory because Go Template can not handle
	// multiple files with the same name in different directories.

	webTemplates := loadAllTemplates(web.TemplateFS)
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
