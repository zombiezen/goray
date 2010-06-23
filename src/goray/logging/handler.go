//
//  goray/logging/handler.go
//  goray
//
//  Created by Ross Light on 2010-06-22.
//

package logging

import (
	"fmt"
	"io"
	"os"
)

/* Handler defines an interface for handling a single record. */
type Handler interface {
	Handle(Record)
}

func shortcut(level int, log Handler, format string, args ...interface{}) {
	if log != nil {
		log.Handle(BasicRecord{fmt.Sprintf(format, args), level})
	}
}

/* VerboseDebug is a shortcut for sending a record to a handler at VerboseDebugLevel. */
func VerboseDebug(log Handler, format string, args ...interface{}) {
	shortcut(VerboseDebugLevel, log, format, args)
}
/* Debug is a shortcut for sending a record to a handler at DebugLevel. */
func Debug(log Handler, format string, args ...interface{}) {
	shortcut(DebugLevel, log, format, args)
}
/* Info is a shortcut for sending a record to a handler at InfoLevel. */
func Info(log Handler, format string, args ...interface{}) {
	shortcut(InfoLevel, log, format, args)
}
/* Warning is a shortcut for sending a record to a handler at WarningLevel. */
func Warning(log Handler, format string, args ...interface{}) {
	shortcut(WarningLevel, log, format, args)
}
/* Error is a shortcut for sending a record to a handler at ErrorLevel. */
func Error(log Handler, format string, args ...interface{}) {
	shortcut(ErrorLevel, log, format, args)
}
/* Critical is a shortcut for sending a record to a handler at CriticalLevel. */
func Critical(log Handler, format string, args ...interface{}) {
	shortcut(CriticalLevel, log, format, args)
}

type writerHandler struct {
	writer io.Writer
	ch     chan Record
	done   chan bool
}

/* NewWriterHandler creates a logging handler that outputs to an io.Writer. */
func NewWriterHandler(w io.Writer) Handler {
	handler := &writerHandler{w, make(chan Record, 10), make(chan bool)}
	go handler.writerTask()
	return handler
}

func (handler *writerHandler) writerTask() {
	for rec := range handler.ch {
		io.WriteString(handler.writer, rec.String()+"\n")
	}
	handler.done <- true
}

func (handler *writerHandler) Handle(rec Record) {
	if handler.ch != nil {
		handler.ch <- rec
	}
}

func (handler *writerHandler) Close() os.Error {
	if handler.ch != nil {
		close(handler.ch)
		<-handler.done
		handler.ch = nil
	}
	return nil
}
