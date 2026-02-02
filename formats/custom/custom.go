// Copyright © 2026 Kube logging authors
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

package custom

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/kube-logging/log-generator/formats/custom/sysloglike"
	"github.com/kube-logging/log-generator/log"
)

var sequence *sync.Map

func init() {
	sequence = &sync.Map{}
}

func Formats() map[string][]string {
	return map[string][]string{
		"sysloglike": SyslogFormatNames(),
	}
}

func getRuntimeTemplateDir() fs.FS {
	templateDir := strings.TrimSpace(os.Getenv("TEMPLATE_DIR"))
	if templateDir == "" {
		return nil
	}

	return os.DirFS(templateDir)
}

func LogFactory(logType string, format string, randomise bool) (log.Log, error) {
	var templates []fs.FS
	if runtimeTemplateDir := getRuntimeTemplateDir(); runtimeTemplateDir != nil {
		templates = append(templates, runtimeTemplateDir)
	}
	templates = append(templates, sysloglike.TemplateFS)

	switch logType {
	case "sysloglike":
		if randomise {
			return sysloglike.NewRandomSyslog(format, sequence, templates)
		} else {
			return sysloglike.NewSyslog(format, sequence, templates)
		}
	default:
		return nil, fmt.Errorf("invalid type %q", logType)
	}
}

func SyslogFormatNames() []string {
	return log.FormatNames(sysloglike.TemplateFS)
}
