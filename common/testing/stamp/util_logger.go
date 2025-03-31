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
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

type (
	logWrapper struct {
		log.Logger
	}
)

func newLogger(logger log.Logger) log.Logger {
	return &logWrapper{Logger: logger}
}

func (l *logWrapper) Debug(msg string, tags ...tag.Tag) {
	l.Logger.Debug(msg+"\n", tags...) // TODO
}

func (l *logWrapper) Info(msg string, tags ...tag.Tag) {
	l.Logger.Info(msg+"\n", tags...)
}

func (l *logWrapper) Warn(msg string, tags ...tag.Tag) {
	l.Logger.Warn(msg+"\n", tags...)
}

func (l *logWrapper) Error(msg string, tags ...tag.Tag) {
	l.Logger.Error(redStr(msg)+"\n", tags...)
}

func (l *logWrapper) DPanic(msg string, tags ...tag.Tag) {
	l.Logger.DPanic(msg+"\n", tags...)
}

func (l *logWrapper) Panic(msg string, tags ...tag.Tag) {
	l.Logger.Panic(msg+"\n", tags...)
}

func (l *logWrapper) Fatal(msg string, tags ...tag.Tag) {
	msg = "💥 " + redStr(msg)
	l.Logger.Fatal(msg+"\n", tags...)
	panic(msg) // `Fatalf` doesn't seem to stop the test execution
}
