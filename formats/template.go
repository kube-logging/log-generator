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
