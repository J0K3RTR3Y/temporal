// The MIT License
//
// Copyright (c) 2025 Temporal Technologies Inc.  All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package stamp

import (
	"fmt"
	"strings"
	"sync"
)

type (
	evalContext struct {
		Recorder modelRecorder
	}
	modelRecorder struct {
		records []Record
		lock    sync.RWMutex
	}
	Record any
)

func newEvalContext(records ...Record) *evalContext {
	return &evalContext{
		Recorder: modelRecorder{records: records},
	}
}

func (l *modelRecorder) Add(records ...Record) *modelRecorder {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.records = append(l.records, records...)
	return l
}

func (l *modelRecorder) Get() []Record {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.records
}

func (e *evalContext) String() string {
	var sb strings.Builder
	sb.WriteString("EvalContext [")
	records := e.Recorder.Get()
	if len(records) > 0 {
		sb.WriteString("\n")
		for _, r := range records {
			sb.WriteString(fmt.Sprintf("  - %#v\n", r))
		}
	}
	sb.WriteString("]")
	return sb.String()
}
