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

// Package log provides an interface for multi-level logging.
package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// A type that implements Logger can output multi-level log messages.
type Logger interface {
	// Debugf formats its arguments according to the format, analogous to fmt.Printf,
	// and records the text as a log message at Debug level.
	Debugf(format string, args ...interface{})

	// Infof is like Debugf, but at Info level.
	Infof(format string, args ...interface{})

	// Warningf is like Debugf, but at Warning level.
	Warningf(format string, args ...interface{})

	// Errorf is like Debugf, but at Error level.
	Errorf(format string, args ...interface{})

	// Criticalf is like Debugf, but at Critical level.
	Criticalf(format string, args ...interface{})
}

// A writerLog is a logger that sends its messages to a writer.
type writerLog struct {
	w   io.Writer
	buf []byte
	sync.Mutex
}

// New creates a logger that serializes writes to w.
func New(w io.Writer) Logger {
	return &writerLog{w: w}
}

func (l *writerLog) output(s string) error {
	l.Lock()
	defer l.Unlock()
	l.buf = l.buf[:0]
	l.buf = append(l.buf, s...)
	if len(s) > 0 && s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.w.Write(l.buf)
	return err
}

func (l *writerLog) Debugf(format string, args ...interface{}) {
	l.output("DEBUG: " + fmt.Sprintf(format, args...))
}

func (l *writerLog) Infof(format string, args ...interface{}) {
	l.output("INFO: " + fmt.Sprintf(format, args...))
}

func (l *writerLog) Warningf(format string, args ...interface{}) {
	l.output("WARNING: " + fmt.Sprintf(format, args...))
}

func (l *writerLog) Errorf(format string, args ...interface{}) {
	l.output("ERROR: " + fmt.Sprintf(format, args...))
}

func (l *writerLog) Criticalf(format string, args ...interface{}) {
	l.output("CRITICAL: " + fmt.Sprintf(format, args...))
}

// Default logger
var Default Logger = New(os.Stderr)

// Debugf calls Debugf on the default logger. Arguments are handled in the
// manner of fmt.Printf.
func Debugf(format string, args ...interface{}) {
	Default.Debugf(format, args...)
}

// Infof calls Infof on the default logger. Arguments are handled in the
// manner of fmt.Printf.
func Infof(format string, args ...interface{}) {
	Default.Infof(format, args...)
}

// Warningf calls Warningf on the default logger. Arguments are handled in the
// manner of fmt.Printf.
func Warningf(format string, args ...interface{}) {
	Default.Warningf(format, args...)
}

// Errorf calls Errorf on the default logger. Arguments are handled in the
// manner of fmt.Printf.
func Errorf(format string, args ...interface{}) {
	Default.Errorf(format, args...)
}

// Criticalf calls Criticalf on the default logger. Arguments are handled in the
// manner of fmt.Printf.
func Criticalf(format string, args ...interface{}) {
	Default.Criticalf(format, args...)
}
