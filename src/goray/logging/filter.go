//
//  goray/logging/filter.go
//  goray
//
//  Created by Ross Light on 2010-06-22.
//

package logging

import (
	"io"
	"os"
)

/*
   FilterFunc defines a function that operates on a record.

   Possible operations include modifying the record as it passes through,
   returning nil to prevent the record's propagation, or even just passing the
   record through as-is.
*/
type FilterFunc func(Record) Record

/*
   Filter defines a Handler that runs records through a function before passing
   them to another Handler.
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
func NewMinLevelFilter(minLevel Level, handler Handler) Filter {
	return Filter{handler, func(rec Record) Record {
		if rec.Level() < minLevel {
			return nil
		}
		return rec
	}}
}
