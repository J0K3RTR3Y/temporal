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
	"context"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
)

type (
	Scenario struct {
		t           *testing.T
		timeout     time.Duration
		ctx         context.Context
		ctxCancelFn context.CancelFunc
	}
	scenarioOption func(*Scenario)
)

// TODO: apply timeout to scenario or WaitUntil
func WithTimeout(timeout time.Duration) scenarioOption {
	return func(s *Scenario) {
		s.timeout = timeout
	}
}

func DescribeScenario(
	t *testing.T,
	name string,
	f func(t *Scenario),
	opts ...scenarioOption,
) {
	for _, opt := range opts {
		opt(newScenario(t))
	}

	t.Run(name, func(t *testing.T) {
		// always run scenarios in parallel by default
		t.Parallel()

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Scenario panic: %v\n%s", r, debug.Stack())
			}
		}()

		f(newScenario(t))
	})
}

func newScenario(t *testing.T) *Scenario {
	timeout := 10 * time.Second
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	return &Scenario{
		t:           t,
		timeout:     timeout, // TODO: add option to control timeout
		ctx:         ctx,
		ctxCancelFn: cancelFunc,
	}
}

// TODO: Wait options to customize timeout, polling interval, etc.
func (s *Scenario) WaitUntil(f func() Prop[bool]) {
	prop := f()
	s.t.Helper()

	waitTimeout := 2 * time.Second
	waitTimeoutCh := time.After(waitTimeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var lastEvalCtx *evalContext
	for {
		select {
		case <-s.ctx.Done():
			panic(fmt.Sprintf("Scenario timed out after %s: %v", s.timeout, s.ctx.Err()))
		case <-waitTimeoutCh:
			pc := runtime.FuncForPC(reflect.ValueOf(f).Pointer()) // TODO: better formatting
			panic(fmt.Sprintf(`WaitUntil timed out after %s

%s:
%s

%s:
%s
`,
				waitTimeout,
				underlineStr("Property"),
				pc.Name(),
				underlineStr("Input"),
				lastEvalCtx))
		case <-ticker.C:
			res, evalCtx, err := prop.Get()
			if err != nil {
				panic(err)
			}
			if res == true {
				return
			}
			lastEvalCtx = evalCtx
		}
	}
}
