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
// limitations under the License.

package writers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	logger "github.com/sirupsen/logrus"

	"github.com/kube-logging/log-generator/log"
	"github.com/kube-logging/log-generator/metrics"
)

type FileLogWriterConfig struct {
	Path           string
	Create         bool
	Append         bool
	DirMode        os.FileMode
	FileMode       os.FileMode
	SyncAfterWrite bool
}

type FileLogWriter struct {
	config FileLogWriterConfig
	file   *os.File
	mu     sync.Mutex
	closed bool
}

func NewFileWriter(config FileLogWriterConfig) LogWriter {
	flw := &FileLogWriter{
		config: config,
	}
	if err := flw.openLocked(); err != nil {
		logger.Fatalf("failed to open log file: %v", err)
	}

	return flw
}

func (flw *FileLogWriter) Send(l log.Log) {
	msg, size := l.String()
	if l.IsFramed() {
		msg = fmt.Sprintf("%d %s", len(msg), msg)
	}
	msg += "\n"

	flw.mu.Lock()
	defer flw.mu.Unlock()

	if flw.closed {
		logger.Warn("Attempted to write to closed FileLogWriter")
		return
	}

	if flw.wasRotated() {
		logger.Info("Log file was rotated, reopening...")
		if err := flw.openLocked(); err != nil {
			logger.Errorf("failed to reopen rotated log file: %v", err)
			return
		}
	}

	_, err := flw.file.WriteString(msg)
	if err != nil {
		logger.Errorf("error writing to file %s: %v", flw.config.Path, err)
		return
	}

	if flw.config.SyncAfterWrite {
		if err := flw.file.Sync(); err != nil {
			logger.Errorf("error syncing file %s: %v", flw.config.Path, err)
		}
	}

	metrics.EventEmitted.With(l.Labels()).Inc()
	metrics.EventEmittedBytes.With(l.Labels()).Add(size)
}

func (flw *FileLogWriter) Close() {
	flw.mu.Lock()
	defer flw.mu.Unlock()

	if flw.closed {
		return
	}
	flw.closed = true

	if flw.file != nil {
		if err := flw.file.Sync(); err != nil {
			logger.Errorf("error syncing file %s on close: %v", flw.config.Path, err)
		}
		if err := flw.file.Close(); err != nil {
			logger.Errorf("error closing file %s: %v", flw.config.Path, err)
		}
		flw.file = nil
	}
}

func (flw *FileLogWriter) openLocked() error {
	if flw.file != nil {
		_ = flw.file.Sync()
		_ = flw.file.Close()
		flw.file = nil
	}

	dir := filepath.Dir(flw.config.Path)
	if dir != "" && dir != "." {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if !flw.config.Create {
				return fmt.Errorf("directory %s does not exist and file.create is false", dir)
			}
			if err := os.MkdirAll(dir, flw.config.DirMode); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			logger.Infof("Created directory: %s", dir)
		}
	}

	flags := os.O_WRONLY
	if flw.config.Append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	if _, err := os.Stat(flw.config.Path); os.IsNotExist(err) {
		if !flw.config.Create {
			return fmt.Errorf("file %s does not exist and file.create is false", flw.config.Path)
		}
		flags |= os.O_CREATE
	}

	file, err := os.OpenFile(flw.config.Path, flags, flw.config.FileMode)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", flw.config.Path, err)
	}
	flw.file = file

	logger.Infof("Opened log file: %s (append=%v, sync=%v)", flw.config.Path, flw.config.Append, flw.config.SyncAfterWrite)
	return nil
}

func (flw *FileLogWriter) wasRotated() bool {
	if flw.file == nil {
		return true
	}

	currentStat, err := flw.file.Stat()
	if err != nil {
		return true
	}

	pathStat, err := os.Stat(flw.config.Path)
	if err != nil {
		return true
	}

	// compare inodes
	return !os.SameFile(currentStat, pathStat)
}
