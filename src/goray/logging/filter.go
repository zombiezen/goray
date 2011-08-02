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
	"time"
)

// FilterFunc defines a function that operates on a record.
//
// Possible operations include modifying the record as it passes through,
// returning nil to prevent the record's propagation, or even just passing the
// record through as-is.
type FilterFunc func(Record) Record

// Filter defines a Handler that runs records through a function before passing
// them to another Handler.
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

// NewMinLevelFilter creates a new Filter that removes records that are below a certain level.
func NewMinLevelFilter(handler Handler, minLevel Level) Filter {
	return Filter{handler, func(rec Record) Record {
		if rec.Level() < minLevel {
			return nil
		}
		return rec
	}}
}

// FormatFunc defines a function that takes a record and returns a formatted
// message.
//
// Along with NewFormatFilter, this provides a simple way to format your log
// messages.
type FormatFunc func(Record) string

// NewFormatFilter creates a new Filter that formats records that pass through it.
func NewFormatFilter(handler Handler, f FormatFunc) Filter {
	return Filter{handler, func(rec Record) Record {
		return NewFormattedRecord(rec, f(rec))
	}}
}

// DefaultFormatter returns a Filter that formats a record into a reasonable log string.
func DefaultFormatter(handler Handler) Filter {
	return NewFormatFilter(handler, func(rec Record) string {
		switch orig := GetUnformattedRecord(rec).(type) {
		case DatedRecord:
			return fmt.Sprintf(
				"[%s]%v:%s",
				orig.Timestamp().Format(time.RFC3339),
				rec.Level(),
				rec.String(),
			)
		}
		return fmt.Sprintf("%v:%s", rec.Level(), rec.String())
	})
}
