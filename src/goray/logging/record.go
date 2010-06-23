//
//  goray/logging/record.go
//  goray
//
//  Created by Ross Light on 2010-06-22.
//

package logging

import "fmt"

// Predefined log levels
const (
	VerboseDebugLevel = (iota + 1) * 10
	DebugLevel
	InfoLevel
	WarningLevel
	ErrorLevel
	CriticalLevel
)

/* Level is an integer representing the severity of a record. */
type Level int

func (lvl Level) String() string {
	switch lvl {
	case VerboseDebugLevel:
		return "Verbose"
	case DebugLevel:
		return "Debug"
	case InfoLevel:
		return "Info"
	case WarningLevel:
		return "Warning"
	case ErrorLevel:
		return "Error"
	case CriticalLevel:
		return "Critical"
	}
	return fmt.Sprintf("Level%d", int(lvl))
}

/* Record defines a simple log record. */
type Record interface {
	Level() Level
	String() string
}

/* StringRecord is a simple, info-level record. */
type StringRecord string

func (rec StringRecord) Level() Level   { return InfoLevel }
func (rec StringRecord) String() string { return string(rec) }

/* BasicRecord stores a message and a level. */
type BasicRecord struct {
	Message     string
	RecordLevel Level
}

func (rec BasicRecord) Level() Level   { return rec.RecordLevel }
func (rec BasicRecord) String() string { return rec.Message }

/* A FormattedRecord wraps another record with a customized string. */
type FormattedRecord struct {
	Original Record
	Message  string
}

func NewFormattedRecord(orig Record, msg string) FormattedRecord {
	// Unwrap original record if we were given another formatted record.
	if fmtRec, ok := orig.(FormattedRecord); ok {
		orig = fmtRec.Original
	}
	// Create record
	return FormattedRecord{orig, msg}
}

func (rec FormattedRecord) Level() Level   { return rec.Original.Level() }
func (rec FormattedRecord) String() string { return rec.Message }
