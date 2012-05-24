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

/*
   Package logging is a logging system similar to Python's.

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

// MainLog is the global logger object for the program.
// You do not have to use it if you don't want to.
var MainLog = NewLogger()

// Logger dispatches records to a set of handlers.
type Logger struct {
	handlers []Handler
	ch       chan Record
}

// NewLogger creates a Logger object without any handlers attached.
func NewLogger() (log *Logger) {
	return &Logger{handlers: make([]Handler, 0)}
}

// AddHandler adds a new handler to the logger.
func (log *Logger) AddHandler(handler Handler) {
	log.handlers = append(log.handlers, handler)
}

// Log creates a new BasicRecord and sends it to the handlers.
func (log *Logger) Log(level Level, message string) {
	shortcut(level, log, "%s", message)
}

// Logf creates a new BasicRecord from a Printf format string and sends it to the handlers.
func (log *Logger) Logf(level Level, format string, args ...interface{}) {
	shortcut(level, log, format, args...)
}

// VerboseDebug is a shortcut for Logf(VerboseDebugLevel).
func (log *Logger) VerboseDebug(format string, args ...interface{}) {
	VerboseDebug(log, format, args...)
}

// Debug is a shortcut for Logf(DebugLevel).
func (log *Logger) Debug(format string, args ...interface{}) {
	Debug(log, format, args...)
}

// Info is a shortcut for Logf(InfoLevel).
func (log *Logger) Info(format string, args ...interface{}) {
	Info(log, format, args...)
}

// Warning is a shortcut for Logf(WarningLevel).
func (log *Logger) Warning(format string, args ...interface{}) {
	Warning(log, format, args...)
}

// Error is a shortcut for Logf(ErrorLevel).
func (log *Logger) Error(format string, args ...interface{}) {
	Error(log, format, args...)
}

// Critical is a shortcut for Logf(CriticalLevel).
func (log *Logger) Critical(format string, args ...interface{}) {
	Critical(log, format, args...)
}

// Handle dispatches a record to the logger's handlers.
func (log *Logger) Handle(rec Record) {
	for _, h := range log.handlers {
		h.Handle(rec)
	}
}
