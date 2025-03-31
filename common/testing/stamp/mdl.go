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
	"hash/fnv"
	"reflect"
)

type (
	Model[T modelWrapper, A action] struct {
		*internalModel
	}
	internalModel struct { // model representation without type annotations
		id       modelID
		domainID ID
		env      modelEnv
		typeOf   modelType
		evalCtx  *evalContext
		genCache map[any]any
	}
	modelType struct {
		ptrType    reflect.Type
		structType reflect.Type
		name       string
	}
	modelIdent struct {
		ty modelType
		id ID
	}
	ModelAccessor interface {
		getModel() *internalModel // read-only!
		GetID() ID
	}
	modelWrapper interface { // API for the user-defined struct
		ModelAccessor
		record(...Record)
		str() string
		getID() modelID
		setScope(modelWrapper)
		getScope() modelWrapper
		getType() modelType
		getEvalCtx() *evalContext
		setModel(*internalModel)
		setDomainID(id ID)
	}
	Scope[T modelWrapper] struct {
		mw T
	}
	identityHandler[A any] interface {
		Identity(A) ID
	}
	actionHandler[A any] interface {
		OnAction(A) Record
	}
	ID string
)

func newModel[A action](env *ModelEnv[A], mdlType modelType, scope modelWrapper) modelWrapper {
	mw := reflect.New(mdlType.structType).Interface().(modelWrapper)
	mw.setModel(
		&internalModel{
			env:     env,
			id:      env.nextID(),
			typeOf:  mdlType,
			evalCtx: newEvalContext(),
		})
	mw.setScope(scope)
	return mw
}

// Note: not `String` to avoid concurrency issues when being printed by testing goroutine
func (m *internalModel) str() string {
	return fmt.Sprintf("%s[%s]", m.typeOf.name, m.getDomainID())
}

func (m *Model[T, A]) setModel(mdl *internalModel) {
	m.internalModel = mdl
}

func (m *Model[T, A]) getModelAccessor() ModelAccessor {
	return m.getModel()
}

func (m *internalModel) Seed() int {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%s_%d", m.typeOf.String(), m.env.Seed())))
	return int(h.Sum32())
}

func (m *internalModel) getEnv() modelEnv {
	return m.env
}

func (m *internalModel) getType() modelType {
	return m.typeOf
}

func (m *internalModel) getDomainID() ID {
	return m.domainID
}

func (m *internalModel) setDomainID(id ID) {
	m.domainID = id
}

// GetID returns the domain ID of the model.
func (m *internalModel) GetID() ID {
	return m.getDomainID()
}

func (m *internalModel) getID() modelID {
	if m.id == 0 {
		m.getEnv().Fatal(fmt.Sprintf("%q is missing an ID", m.typeOf))
	}
	return m.id
}

func (m *internalModel) getModel() *internalModel {
	return m
}

func (m *internalModel) getEvalCtx() *evalContext {
	return m.evalCtx
}

func (m *internalModel) setModel(_ *internalModel) {
	panic("not implemented")
}

func (m *internalModel) is_model() {
	panic("not implemented")
}

func (m *internalModel) record(records ...Record) {
	m.evalCtx.Recorder.Add(records...)
}

func (id ID) Gen() Gen[ID] {
	return GenName[ID]()
}

func (s Scope[T]) getScope() modelWrapper {
	return s.mw
}

func (s Scope[T]) GetScope() T {
	return s.mw
}

func (s *Scope[T]) setScope(mw modelWrapper) {
	s.mw = mw.(T)
}

func (t *modelType) String() string {
	return t.name
}
