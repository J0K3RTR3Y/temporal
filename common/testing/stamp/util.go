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
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	goValidator "github.com/go-playground/validator/v10"
)

var (
	boldStr      = color.New(color.Bold).SprintFunc()
	redStr       = color.New(color.FgRed).SprintFunc()
	underlineStr = color.New(color.Underline).SprintFunc()
	simpleSpew   = spew.NewDefaultConfig()
	validator    = goValidator.New()
)

func init() {
	color.NoColor = false

	simpleSpew.DisablePointerAddresses = true
	simpleSpew.DisableCapacities = true
	simpleSpew.DisableMethods = true
	simpleSpew.MaxDepth = 2
}

func qualifiedTypeName(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}
