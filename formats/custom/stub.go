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

package custom

import (
	"github.com/kube-logging/log-generator/log"
)

// Formats returns supported template formats for a given type
func Formats() map[string][]string {
	return map[string][]string{}
}

// LogFactory creates log events for a log format and optionally randomises it
func LogFactory(logType string, format string, randomise bool) (log.Log, error) {
	return nil, nil
}
