/*
	Copyright (c) 2011 Ross Light.
	Copyright (c) 2005 Mathias Wein, Alejandro Conty, and Alfredo de Greef.

	This file is part of goray.

	goray is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	goray is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with goray.  If not, see <http://www.gnu.org/licenses/>.
*/

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
	now := time.Now().UTC()
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
	writer io.Writer
	mu     sync.Mutex
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

// A CircularHandler keeps an in-memory rotating log.
type CircularHandler struct {
	records       []Record
	tail          int
	started, full bool
	mu            sync.RWMutex
}

func NewCircularHandler(n int) *CircularHandler {
	return &CircularHandler{records: make([]Record, n)}
}

func (handler *CircularHandler) Init(n int) {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	// TODO: Reuse memory?
	handler.records = make([]Record, n)
	handler.tail = 0
	handler.started, handler.full = false, false
}

func (handler *CircularHandler) Len() int {
	handler.mu.RLock()
	defer handler.mu.RUnlock()
	switch {
	case !handler.started:
		return 0
	case handler.full:
		return len(handler.records)
	}
	return handler.tail
}

func (handler *CircularHandler) Cap() int {
	handler.mu.RLock()
	defer handler.mu.RUnlock()
	return len(handler.records)
}

func (handler *CircularHandler) Handle(rec Record) {
	handler.mu.Lock()
	defer handler.mu.Unlock()
	handler.started = true
	handler.records[handler.tail] = rec
	handler.tail = (handler.tail + 1) % len(handler.records)
	if handler.tail == 0 {
		handler.full = true
	}
}

func (handler *CircularHandler) Full() bool {
	handler.mu.RLock()
	defer handler.mu.RUnlock()
	return handler.full
}

func (handler *CircularHandler) Records() (recs []Record) {
	handler.mu.RLock()
	defer handler.mu.RUnlock()
	switch {
	case !handler.started:
		recs = []Record{}
	case handler.full:
		n := len(handler.records)
		recs = make([]Record, n)
		copy(recs, handler.records[handler.tail:])
		copy(recs[n-handler.tail:], handler.records[:handler.tail])
	default:
		recs = make([]Record, handler.tail)
		copy(recs, handler.records[:handler.tail])
	}
	return
}
