// Copyright (c) 2023 Cisco All Rights Reserved.

package formats

import (
	"strings"
	"testing"
	"text/template"

	"github.com/banzaicloud/log-generator/formats/syslog"
)

func TestSyslogFormats(t *testing.T) {
	// Separate Template tree per directory because Go Template can not handle
	// multiple files with the same name in different directories.

	syslogTemplates := loadAllTemplates(syslog.TemplateFS)
	assertFormatAll(t, syslogTemplates)
}

func assertFormatAll(t *testing.T, templates []*template.Template) {
	for _, f := range templates {
		format := strings.TrimSuffix(f.Name(), ".tmpl")
		log, err := NewSyslog(format)
		if err != nil {
			t.Fatalf("Failed to create log, format=%q, %v", format, err)
		}

		if l, _ := log.String(); len(strings.TrimSpace(l)) == 0 {
			t.Logf("Rendered log is empty, format=%q", format)
		}
	}
}
