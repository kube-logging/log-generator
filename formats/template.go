// Copyright (c) 2023 Cisco All Rights Reserved.

package formats

import (
	"fmt"
	"io/fs"
	"strings"
	"text/template"
)

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

func loadAllTemplates(fs fs.FS) []*template.Template {
	return template.Must(template.ParseFS(fs, "*.tmpl")).Templates()
}
