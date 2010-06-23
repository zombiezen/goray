//
//  goray/logging/record.go
//  goray
//
//  Created by Ross Light on 2010-06-22.
//

package logging

// Predefined log levels
const (
	VerboseDebugLevel = (iota + 1) * 10
	DebugLevel
	InfoLevel
	WarningLevel
	ErrorLevel
	CriticalLevel
)

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
