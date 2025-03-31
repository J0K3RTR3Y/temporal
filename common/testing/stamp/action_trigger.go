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
// THE SOFTMARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package stamp

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

var (
	DebugTriggerID = TriggerID("<debugTriggerID>")
)

type (
	action interface {
		String() string
	}
	actionWrapper struct {
		triggerID TriggerID
		action    any
	}
	TriggerID                      ID
	TriggerActor[MA ModelAccessor] struct {
		actor MA
	}
	ActorModel[MA ModelAccessor] struct {
		actor MA
	}
	actorModel[MA ModelAccessor] interface {
		ModelAccessor
		getModelAccessor() MA
		OnTrigger(TriggerID, any) // implemented by user
	}
	TriggerTarget[M ModelAccessor] struct {
		mdl M
	}
	trigger[A actorModel[AMA], AMA ModelAccessor, T ModelAccessor, P payload] interface {
		Get(GenContext) P
		GetActor() AMA
		getTarget() T
	}
	payload           interface{}
	triggerOption     interface{} // TODO: add marker
	triggerIdDebugOpt struct {
		id TriggerID
	}
)

func NewActorModel[AM ModelAccessor](mdl AM) ActorModel[AM] {
	return ActorModel[AM]{actor: mdl}
}

func WithDebugTriggerID() triggerOption {
	if _, ok := os.LookupEnv("CI"); ok {
		panic("WithDebugTriggerID is only available in CI, only for local debugging")
	}
	return triggerIdDebugOpt{id: DebugTriggerID}
}

func Trigger[
	A actorModel[AMA], // the actor that is triggering the action
	AMA ModelAccessor, // the underlying model of the actor
	T ModelAccessor, // the target model that is being triggered
	P payload, // the payload of the trigger
](
	actor A,
	trg trigger[A, AMA, T, P],
	opts ...triggerOption,
) T {
	env := actor.getModel().getEnv()

	triggerID := TriggerID(uuid.NewString())
	for _, opt := range opts {
		if debugID, ok := opt.(triggerIdDebugOpt); ok {
			triggerID = debugID.id
		}
	}

	// validate trigger
	if err := validator.Struct(trg); err != nil {
		env.Fatal(fmt.Sprintf("trigger '%s' is invalid: %v", trg, err))
	}

	// copy trigger to assign actor
	// TODO: check that it wasn't set by user
	copyVal := reflect.New(reflect.TypeOf(trg))
	copyVal.Elem().Set(reflect.ValueOf(trg))
	copyVal.Elem().FieldByName("TriggerActor").
		Set(reflect.ValueOf(TriggerActor[AMA]{actor: actor.getModelAccessor()}))
	newTrigger := copyVal.Elem().Interface().(trigger[A, AMA, T, P])

	// setup trigger callback
	var res T
	var triggered bool
	matchItems := map[*actionWrapper][]modelWrapper{}
	matchSet := map[*actionWrapper]map[modelID]struct{}{}
	matchTypes := map[modelType]struct{}{}
	env.onMatched(triggerID, func(act *actionWrapper, mdl modelWrapper) {
		if triggered {
			// already triggered, ignore
			return
		}

		// keep track of matched models for error reporting
		if _, ok := matchSet[act]; !ok {
			matchSet[act] = map[modelID]struct{}{}
			matchItems[act] = []modelWrapper{}
		}
		if _, ok := matchSet[act][mdl.getID()]; !ok {
			matchSet[act][mdl.getID()] = struct{}{}
			matchItems[act] = append(matchItems[act], mdl)
		}
		matchTypes[mdl.getType()] = struct{}{}

		if matchedMdl, ok := mdl.(T); ok {
			triggered = true
			res = matchedMdl
		}
	})

	// send trigger
	triggerPayload := newTrigger.Get(env)
	env.Info(fmt.Sprintf("Triggering '%v'", boldStr(triggerPayload)))
	actor.OnTrigger(triggerID, triggerPayload)

	// verify that the trigger arrived at the target
	if !triggered {
		var matchedDetails strings.Builder
		for a, models := range matchItems {
			matchedDetails.WriteString("- ")
			matchedDetails.WriteString(a.String())
			matchedDetails.WriteString(":\n")
			for _, mdl := range models {
				matchedDetails.WriteString("\t- ")
				matchedDetails.WriteString(mdl.str())
				matchedDetails.WriteString("\n")
			}
		}

		var expectedNextMatch string
		var notMatched strings.Builder
		triggerTargetType := reflect.TypeFor[T]().Elem()
		for _, mdlType := range env.pathTo(triggerTargetType) {
			if _, ok := matchTypes[mdlType]; ok {
				continue
			}
			if expectedNextMatch == "" {
				expectedNextMatch = mdlType.name
			}
			notMatched.WriteString("- ")
			notMatched.WriteString(mdlType.String())
			notMatched.WriteString("\n")
		}

		env.Fatal(fmt.Sprintf(`
Trigger '%s' by actor '%s' did not arrive at target '%s'.

%s:
- check the identity handler of '%s' since it was expected to match next

%s:
%s
%s:
%s
%s:
%s`,
			boldStr(reflect.TypeOf(trg).String()),
			boldStr(reflect.TypeOf(actor).String()),
			boldStr(reflect.TypeOf(trg.getTarget()).String()),
			underlineStr("Hint"),
			boldStr(expectedNextMatch),
			underlineStr("Not Matched"),
			notMatched.String(),
			underlineStr("Matched"),
			matchedDetails.String(),
			underlineStr("Trigger"),
			simpleSpew.Sdump(newTrigger)))
	}

	return res
}
func newActionWrapper[A action](triggerID TriggerID, action A) *actionWrapper {
	return &actionWrapper{
		triggerID: triggerID,
		action:    action,
	}
}

func (a *actionWrapper) String() string {
	return fmt.Sprintf("%s", a.action)
}

func (tt TriggerTarget[M]) getTarget() M {
	return tt.mdl
}

func (ta TriggerActor[A]) GetActor() A {
	return ta.actor
}

func (a *ActorModel[MA]) getModel() *internalModel {
	return a.actor.getModel()
}

func (a *ActorModel[MA]) GetID() ID {
	return a.actor.GetID()
}

func (a *ActorModel[MA]) getModelAccessor() MA {
	return a.actor
}
