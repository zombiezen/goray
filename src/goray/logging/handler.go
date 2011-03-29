//
//	goray/logging/handler.go
//	goray
//
//	Created by Ross Light on 2010-06-22.
//

package logging

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Handler defines an interface for handling a single record.
type Handler interface {
	Handle(Record)
}

func shortcut(level Level, log Handler, format string, args ...interface{}) {
	now := time.UTC()
	if log != nil {
		log.Handle(BasicRecord{fmt.Sprintf(format, args...), level, now})
	}
}

// VerboseDebug is a shortcut for sending a record to a handler at VerboseDebugLevel.
func VerboseDebug(log Handler, format string, args ...interface{}) {
	shortcut(VerboseDebugLevel, log, format, args...)
}
// Debug is a shortcut for sending a record to a handler at DebugLevel.
func Debug(log Handler, format string, args ...interface{}) {
	shortcut(DebugLevel, log, format, args...)
}
// Info is a shortcut for sending a record to a handler at InfoLevel.
func Info(log Handler, format string, args ...interface{}) {
	shortcut(InfoLevel, log, format, args...)
}
// Warning is a shortcut for sending a record to a handler at WarningLevel.
func Warning(log Handler, format string, args ...interface{}) {
	shortcut(WarningLevel, log, format, args...)
}
// Error is a shortcut for sending a record to a handler at ErrorLevel.
func Error(log Handler, format string, args ...interface{}) {
	shortcut(ErrorLevel, log, format, args...)
}
// Critical is a shortcut for sending a record to a handler at CriticalLevel.
func Critical(log Handler, format string, args ...interface{}) {
	shortcut(CriticalLevel, log, format, args...)
}

type writerHandler struct {
	mu     sync.Mutex
	writer io.Writer
}

// NewWriterHandler creates a logging handler that outputs to an io.Writer.
func NewWriterHandler(w io.Writer) Handler {
	return &writerHandler{writer: w}
}

func (handler *writerHandler) Handle(rec Record) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	io.WriteString(handler.writer, rec.String()+"\n")
}
