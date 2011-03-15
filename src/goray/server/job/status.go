//
//  goray/server/job/status.go
//  goray
//
//  Created by Ross Light on 2011-03-14.
//

package job

import (
	"fmt"
	"os"
	"goray/time"
)

type Status struct {
	Code  StatusCode
	Error os.Error

	ReadTime   time.Time
	UpdateTime time.Time
	RenderTime time.Time
	WriteTime  time.Time
}

func (status Status) String() string {
	switch status.Code {
	case StatusDone:
		return fmt.Sprintf("%v (%v)", status.Code, status.TotalTime())
	case StatusError:
		return fmt.Sprintf("%v: %v", status.Code, status.Error)
	}
	return status.Code.String()
}

func (status Status) TotalTime() time.Time {
	return status.ReadTime + status.UpdateTime + status.RenderTime + status.WriteTime
}

func (status Status) Started() bool  { return status.Code.Started() }
func (status Status) Finished() bool { return status.Code.Finished() }

type StatusCode int

const (
	StatusNew StatusCode = 0
	StatusDone = 200
	StatusRendering = 100
	StatusError = 500
)

func (code StatusCode) String() string {
	switch code {
	case StatusNew:
		return "New"
	case StatusDone:
		return "Done"
	case StatusRendering:
		return "Rendering"
	case StatusError:
		return "Failed"
	}
	return fmt.Sprintf("%d", int(code))
}

func (code StatusCode) GoString() string {
	switch code {
	case StatusNew:
		return "job.StatusNew"
	case StatusDone:
		return "job.StatusDone"
	case StatusRendering:
		return "job.StatusRendering"
	case StatusError:
		return "job.StatusError"
	}
	return fmt.Sprintf("job.StatusCode(%d)", int(code))
}

func (code StatusCode) Started() bool {
	return code != StatusNew
}

func (code StatusCode) Finished() bool {
	return code == StatusDone || code == StatusError
}
