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

package writers

import (
	"fmt"

	"github.com/kube-logging/log-generator/log"
	"github.com/kube-logging/log-generator/metrics"
)

type StdoutLogWriter struct{}

func NewStdoutWriter() LogWriter {
	return &StdoutLogWriter{}
}

func (*StdoutLogWriter) Send(l log.Log) {
	msg, size := l.String()

	if l.IsFramed() {
		msg = fmt.Sprintf("%d %s", len(msg), msg)
	}

	fmt.Println(msg)

	metrics.EventEmitted.With(l.Labels()).Inc()
	metrics.EventEmittedBytes.With(l.Labels()).Add(size)
}

func (*StdoutLogWriter) Close() {}
