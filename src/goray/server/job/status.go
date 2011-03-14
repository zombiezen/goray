//
//  goray/server/job/status.go
//  goray
//
//  Created by Ross Light on 2011-03-14.
//

package job

import (
	"fmt"
)

type StatusCode int

const (
	StatusNew StatusCode = iota
	StatusDone
	StatusRendering
	StatusError
)

func (status StatusCode) String() string {
	switch status {
	case StatusNew:
		return "New"
	case StatusDone:
		return "Done"
	case StatusRendering:
		return "Rendering"
	case StatusError:
		return "Failed"
	}
	return fmt.Sprintf("StatusCode(%d)", int(status))
}

func (status StatusCode) GoString() string {
	switch status {
	case StatusNew:
		return "job.StatusNew"
	case StatusDone:
		return "job.StatusDone"
	case StatusRendering:
		return "job.StatusRendering"
	case StatusError:
		return "job.StatusError"
	}
	return fmt.Sprintf("job.StatusCode(%d)", int(status))
}

// Name returns the internal name of the status.
func (status StatusCode) Name() string {
	switch status {
	case StatusNew:
		return "new"
	case StatusDone:
		return "done"
	case StatusRendering:
		return "rendering"
	case StatusError:
		return "error"
	}
	return fmt.Sprint(int(status))
}

func (status StatusCode) Started() bool {
	return status != StatusNew
}

func (status StatusCode) Finished() bool {
	return status == StatusDone || status == StatusError
}
