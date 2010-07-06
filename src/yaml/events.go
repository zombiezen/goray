//
//	yaml/events.go
//	goray
//
//	Created by Ross Light on 2010-06-24.
//

package yaml

const (
    NO_EVENT = iota

    STREAM_START_EVENT
    STREAM_END_EVENT

    DOCUMENT_START_EVENT
    DOCUMENT_END_EVENT

    ALIAS_EVENT
    SCALAR_EVENT

    SEQUENCE_START_EVENT
    SEQUENCE_END_EVENT

    MAPPING_START_EVENT
    MAPPING_END_EVENT
)
type EventKind int

type Event interface {
	GetKind() EventKind
	GetStart() Position
	GetEnd() Position
}
