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
	reflect "reflect"
	"regexp"
	"sync"

	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/softassert"
)

var (
	scopeTypePattern = regexp.MustCompile(`Scope\[(.*)]`)
)

type (
	ModelEnv[A action] struct {
		log.Logger

		root                Root
		assert              *require.Assertions
		lock                sync.Mutex
		modelSet            ModelSet
		currentTick         int
		currentMdlCounter   modelID
		childrenIdx         map[modelID]map[modelIdent]modelWrapper
		callbackLock        sync.RWMutex
		triggerCallbacksIdx map[TriggerID]func(*actionWrapper, modelWrapper)
	}
	modelEnv interface {
		log.Logger
		Seed() int
		onMatched(TriggerID, func(act *actionWrapper, mdl modelWrapper)) func()
		pathTo(reflect.Type) []modelType
	}
	modelID int // internal identifier only - different from public domain ID
)

func NewModelEnv[A action](
	logger log.Logger,
	modelSet ModelSet,
) *ModelEnv[A] {
	res := &ModelEnv[A]{
		Logger:              newLogger(logger),
		modelSet:            modelSet.clone(), // clone to avoid data races
		childrenIdx:         make(map[modelID]map[modelIdent]modelWrapper),
		triggerCallbacksIdx: make(map[TriggerID]func(*actionWrapper, modelWrapper)),
	}
	res.root = Root{env: res}
	return res
}

func (e *ModelEnv[A]) HandleAction(triggerID TriggerID, action A) {
	defer func() {
		// recover from panics to not disturb the test
		if r := recover(); r != nil {
			softassert.Fail(e, fmt.Sprintf("%v", r))
		}
	}()

	e.lock.Lock()
	defer func() {
		e.currentTick += 1
		e.lock.Unlock()
	}()

	aw := newActionWrapper(triggerID, action)

	// update models
	e.walk(&e.root, func(parent, child modelWrapper) bool {
		if child == nil {
			e.createNewChildren(parent, aw)
			return true
		}
		return e.updateExistingChild(parent, child, aw)
	})

	// verify props
	e.walk(&e.root, func(_, child modelWrapper) bool {
		if child == nil {
			return false
		}
		for _, verifyFn := range e.modelSet.propertiesOf(child) {
			p, res, err := verifyFn(child)
			if err != nil {
				e.Error(err.Error())
			}
			if res != true {
				e.Error(propErr{prop: p.String(), err: PropViolated}.Error())
			}
		}
		return true
	})
}

func (e *ModelEnv[A]) handleAction(id TriggerID, action any) {
	e.HandleAction(id, action.(A))
}

func (e *ModelEnv[A]) createNewChildren(
	parent modelWrapper,
	action *actionWrapper,
) {
	for _, mdlType := range e.modelSet.childTypesOf(parent) {
		// check identity (using an example model to not create a new one)
		exampleMdl := e.modelSet.exampleOf(mdlType)
		exampleMdl.setScope(parent)
		id := e.modelSet.identify(exampleMdl, action)
		if id == "" {
			// unable to identify
			continue
		}
		parentModelID := parent.getID()
		mdlIdentity := modelIdent{ty: mdlType, id: id}
		if _, ok := e.childrenIdx[parentModelID][mdlIdentity]; ok {
			// already exists
			continue
		}

		// create new model
		newChild := newModel(e, mdlType, parent)
		if id2 := e.modelSet.identify(newChild, action); id2 != id { // have to run it again for side effects
			e.Fatal(fmt.Sprintf("identity mismatch: %q != %q", id, id2))
		}
		newChild.setDomainID(id)
		e.Info(fmt.Sprintf("%s created by %s", boldStr(newChild.str()), action),
			tag.NewStringTag("parent", parent.str()),
			tag.NewStringTag("triggerID", string(action.triggerID)))

		// index model
		if _, ok := e.childrenIdx[parentModelID]; !ok {
			e.childrenIdx[parentModelID] = make(map[modelIdent]modelWrapper)
		}
		e.childrenIdx[parentModelID][mdlIdentity] = newChild

		// no need to do anything else here!
		// child will be visited next automatically
	}
}

func (e *ModelEnv[A]) updateExistingChild(
	parent, child modelWrapper,
	action *actionWrapper,
) bool {
	// check identity
	id := e.modelSet.identify(child, action)
	if id == "" {
		// unable to identify
		return false
	}
	if id != child.GetID() {
		// not the same model instance
		return false
	}

	e.Debug(fmt.Sprintf("%s matches %v", boldStr(child.str()), action),
		tag.NewStringTag("parent", parent.str()),
		tag.NewStringTag("triggerID", string(action.triggerID)))

	// invoke action handler
	triggered := e.modelSet.trigger(child, action)
	if triggered {
		e.Info(fmt.Sprintf("%s triggered by %s", boldStr(child.str()), action),
			tag.NewStringTag("parent", parent.str()),
			tag.NewStringTag("triggerID", string(action.triggerID)))
	}

	// invoke trigger callback
	if action.triggerID != "" {
		e.callbackLock.RLock()
		if cb, ok := e.triggerCallbacksIdx[action.triggerID]; ok {
			cb(action, child)
		}
		e.callbackLock.RUnlock()
	}

	return true
}

func (e *ModelEnv[A]) walk(start modelWrapper, fn func(modelWrapper, modelWrapper) bool) {
	var walkModelsFn func(modelWrapper)
	walkModelsFn = func(mw modelWrapper) {
		fn(mw, nil) // might create new children

		for _, child := range e.childrenIdx[mw.getID()] {
			matched := fn(mw, child)
			if !matched {
				continue
			}
			walkModelsFn(child)
		}
	}
	walkModelsFn(start)
}

func (e *ModelEnv[A]) nextID() modelID {
	e.currentMdlCounter++
	return e.currentMdlCounter
}

func (e *ModelEnv[A]) pathTo(ty reflect.Type) []modelType {
	return e.modelSet.pathTo(ty)
}

func (e *ModelEnv[A]) Seed() int {
	// TODO: allow providing a seed
	return 0
}

// TODO: with timeout
func (e *ModelEnv[A]) Context() context.Context {
	return context.Background()
}

func (e *ModelEnv[A]) Root() *Root {
	return &e.root
}

func (e *ModelEnv[A]) onMatched(triggerID TriggerID, fn func(*actionWrapper, modelWrapper)) func() {
	e.callbackLock.Lock()
	defer e.callbackLock.Unlock()

	e.triggerCallbacksIdx[triggerID] = fn
	return func() {
		delete(e.triggerCallbacksIdx, triggerID)
	}
}
