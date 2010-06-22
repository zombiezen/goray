//
//  goray/logging.go
//  goray
//
//  Created by Ross Light on 2010-06-07.
//

/*
   The logging package provides a logging system similar to Python's.

   The core part of the logging package is the Handler interface, which simply
   receives a log record.  The power of this package comes in the fact that you
   can chain handlers together.  For example, the minimum level filter only
   allows messages above a certain level to trickle down to the output; the
   filter could be connected directly to an output handler or it may be
   connected to another filter handler.

   Interestingly, the top-level Logger objects themselves implement the Handler
   interface, which means you can chain logs together.  Complicated
   configurations can be accomplished with just a few handlers.
*/
package logging

import (
	"container/vector"
	"fmt"
	"io"
	"os"
)

/* These constants are predefined logging levels. */
const (
	VerboseDebugLevel = (iota + 1) * 10
	DebugLevel
	InfoLevel
	WarningLevel
	ErrorLevel
	CriticalLevel
)

/*
   MainLog is the global logger object for the program.
   You do not have to use it if you don't want to.
*/
var MainLog = NewLogger()

/* Record defines a simple log record. */
type Record interface {
	Level() int
	String() string
}

/* StringRecord is a simple, info-level record. */
type StringRecord string

func (rec StringRecord) Level() int     { return InfoLevel }
func (rec StringRecord) String() string { return string(rec) }

/* BasicRecord stores a message and a level. */
type BasicRecord struct {
	Message      string
	MessageLevel int
}

func (rec BasicRecord) Level() int     { return rec.MessageLevel }
func (rec BasicRecord) String() string { return rec.Message }

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

/* Logger dispatches records to a set of handlers. */
type Logger struct {
	handlers vector.Vector
	ch       chan Record
}

/* NewLogger creates a Logger object without any handlers attached. */
func NewLogger() (log *Logger) {
	return new(Logger)
}

/* AddHandler adds a new handler to the logger. */
func (log *Logger) AddHandler(handler Handler) {
	log.handlers.Push(handler)
}

/* Log creates a new BasicRecord and sends it to the handlers. */
func (log *Logger) Log(level int, message string) {
	log.Handle(BasicRecord{message, level})
}

/* Logf creates a new BasicRecord from a Printf format string and sends it to the handlers. */
func (log *Logger) Logf(level int, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args)
	log.Handle(BasicRecord{message, level})
}

/* VerboseDebug is a shortcut for Logf(VerboseDebugLevel). */
func (log *Logger) VerboseDebug(format string, args ...interface{}) {
	VerboseDebug(log, format, args)
}
/* Debug is a shortcut for Logf(DebugLevel). */
func (log *Logger) Debug(format string, args ...interface{}) {
	Debug(log, format, args)
}
/* Info is a shortcut for Logf(InfoLevel). */
func (log *Logger) Info(format string, args ...interface{}) {
	Info(log, format, args)
}
/* Warning is a shortcut for Logf(WarningLevel). */
func (log *Logger) Warning(format string, args ...interface{}) {
	Warning(log, format, args)
}
/* Error is a shortcut for Logf(ErrorLevel). */
func (log *Logger) Error(format string, args ...interface{}) {
	Error(log, format, args)
}
/* Critical is a shortcut for Logf(CriticalLevel). */
func (log *Logger) Critical(format string, args ...interface{}) {
	Critical(log, format, args)
}

/* Handle dispatches a record to the logger's handlers. */
func (log *Logger) Handle(rec Record) {
	log.handlers.Do(func(h interface{}) {
		h.(Handler).Handle(rec)
	})
}

/* Close tells all of the logger's handlers to close. */
func (log *Logger) Close() os.Error {
	log.handlers.Do(func(h interface{}) {
		closer, ok := h.(io.Closer)
		if ok {
			closer.Close()
		}
	})
	// TODO: Collect all errors
	return nil
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

/*
   FilterFunc defines a function that operates on a record.

   Possible operations include modifying the record as it passes through, returning
   nil to prevent the record's propagation, or even just passing the record through
   as-is.
*/
type FilterFunc func(Record) Record

/*
   Filter defines a Handler that runs records through a function before passing them to
   another Handler.
*/
type Filter struct {
	Handler Handler
	Func    FilterFunc
}

func (filter Filter) Handle(rec Record) {
	newRecord := filter.Func(rec)
	if newRecord != nil {
		filter.Handler.Handle(newRecord)
	}
}

func (filter Filter) Close() os.Error {
	if closer, ok := filter.Handler.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

/* NewMinLevelFilter creates a new Filter that removes records that are below a certain level. */
func NewMinLevelFilter(minLevel int, handler Handler) Filter {
	return Filter{handler, func(rec Record) Record {
		if rec.Level() < minLevel {
			return nil
		}
		return rec
	}}
}
