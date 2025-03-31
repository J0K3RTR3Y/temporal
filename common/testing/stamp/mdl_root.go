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

	"go.temporal.io/server/common/log"
)

var (
	rootID       modelID = -1
	rootPtrType          = reflect.TypeFor[*Root]()
	rootTypeName         = qualifiedTypeName(rootPtrType.Elem())
	rootType             = modelType{
		ptrType:    rootPtrType,
		structType: rootPtrType.Elem(),
		name:       rootTypeName,
	}
)

type Root struct {
	env         modelEnv
	onTriggerFn func(triggerID TriggerID, payload any)
}

func (r Root) GetLogger() log.Logger {
	panic("not supported")
}

func (r Root) getDomainKey() ID {
	return "<root>"
}

func (r Root) setDomainID(_ ID) {
	panic("not supported")
}

func (r Root) getEnv() modelEnv {
	return r.env
}

func (r Root) setModel(_ *internalModel) {
	panic("not supported")
}

func (r Root) str() string {
	return "Root[]"
}

func (r Root) getModel() *internalModel {
	return &internalModel{
		env: r.env,
		// anything else is unavailable
	}
}

func (r Root) getScope() modelWrapper {
	panic("not supported")
}

func (r Root) record(_ ...Record) {
	panic("not supported")
}

func (r Root) getType() modelType {
	return rootType
}

func (r Root) getID() modelID {
	return rootID
}

func (r Root) GetID() ID {
	panic("not supported")
}

func (r Root) getModelAccessor() *Root {
	return &r
}

func (r Root) getEvalCtx() *evalContext {
	panic("not supported")
}

func (r Root) setScope(modelWrapper) {
	panic("not supported")
}

func (r Root) OnTrigger(triggerID TriggerID, payload any) {
	r.onTriggerFn(triggerID, payload)
}

func (r *Root) SetTriggerHandler(fn func(triggerID TriggerID, payload any)) {
	r.onTriggerFn = fn
}
