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

package job

import (
	"fmt"
	"time"
)

type Status struct {
	Code  StatusCode
	Error error

	ReadTime   time.Duration
	UpdateTime time.Duration
	RenderTime time.Duration
	WriteTime  time.Duration
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

func (status Status) TotalTime() time.Duration {
	return status.ReadTime + status.UpdateTime + status.RenderTime + status.WriteTime
}

func (status Status) Started() bool  { return status.Code.Started() }
func (status Status) Finished() bool { return status.Code.Finished() }

type StatusCode int

const (
	StatusNew StatusCode = 0

	StatusRendering = 100
	StatusReading   = 101
	StatusUpdating  = 102
	StatusWriting   = 103

	StatusDone = 200

	StatusError = 500
)

func (code StatusCode) String() string {
	switch code {
	case StatusNew:
		return "New"
	case StatusRendering:
		return "Rendering"
	case StatusReading:
		return "Reading"
	case StatusUpdating:
		return "Updating"
	case StatusWriting:
		return "Writing"
	case StatusDone:
		return "Done"
	case StatusError:
		return "Failed"
	}
	return fmt.Sprintf("%d", int(code))
}

func (code StatusCode) GoString() string {
	switch code {
	case StatusNew:
		return "job.StatusNew"
	case StatusRendering:
		return "job.StatusRendering"
	case StatusReading:
		return "job.StatusReading"
	case StatusUpdating:
		return "job.StatusUpdating"
	case StatusWriting:
		return "job.StatusWriting"
	case StatusDone:
		return "job.StatusDone"
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
