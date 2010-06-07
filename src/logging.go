//
//  logging.go
//  goray
//
//  Created by Ross Light on 2010-06-07.
//

package logging

import (
	"container/vector"
	"fmt"
	"io"
	"reflect"
)

const (
	DebugLevel = (iota + 1) * 10
	InfoLevel
	WarningLevel
	ErrorLevel
	CriticalLevel
)

var MainLog = NewLogger()

type Logger struct {
	handlers vector.Vector
	ch       chan Record
}

func NewLogger() (log *Logger) {
	return new(Logger)
}

func (log *Logger) AddHandler(handler Handler) {
	log.handlers.Push(handler)
}

func sprintv(format string, args []interface{}) string {
	callArgs := []reflect.Value{reflect.NewValue(format), reflect.NewValue(args)}
	sprintf := reflect.NewValue(fmt.Sprintf).(*reflect.FuncValue)
	return sprintf.Call(callArgs)[0].(*reflect.StringValue).Get()
}

func (log *Logger) Log(level int, message string) {
	log.Handle(BasicRecord{message, level})
}

func (log *Logger) Logf(level int, format string, args ...interface{}) {
	message := sprintv(format, args)
	log.Handle(BasicRecord{message, level})
}

func (log *Logger) Handle(rec Record) {
	log.handlers.Do(func(h interface{}) {
		h.(Handler).Handle(rec)
	})
}

type Record interface {
	Level() int
	String() string
}

type StringRecord string

func (rec StringRecord) Level() int     { return InfoLevel }
func (rec StringRecord) String() string { return string(rec) }

type BasicRecord struct {
	Message      string
	MessageLevel int
}

func (rec BasicRecord) Level() int     { return rec.MessageLevel }
func (rec BasicRecord) String() string { return rec.Message }

type Handler interface {
	Handle(Record)
}

type writerHandler struct {
	writer io.Writer
	ch     chan Record
}

func NewWriterHandler(w io.Writer) Handler {
	handler := &writerHandler{w, make(chan Record)}
	go handler.writerTask()
	return handler
}

func (handler *writerHandler) writerTask() {
	for rec := range handler.ch {
		io.WriteString(handler.writer, rec.String()+"\n")
	}
}

func (handler *writerHandler) Handle(rec Record) {
	handler.ch <- rec
}

type minLevelFilter struct {
	level   int
	handler Handler
}

func NewMinLevelFilter(minLevel int, handler Handler) Handler {
	return minLevelFilter{minLevel, handler}
}

func (handler minLevelFilter) Handle(rec Record) {
	if rec.Level() >= handler.level {
		handler.handler.Handle(rec)
	}
}
