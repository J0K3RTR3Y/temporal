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

package testenv

import (
	"fmt"

	commandpb "go.temporal.io/api/command/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/acceptance/model"
	"go.temporal.io/server/acceptance/testenv/trigger"
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/protobuf/proto"
)

type WorkflowWorker struct {
	stamp.ActorModel[*model.WorkflowWorker]
	env     *testEnv
	handler func() trigger.WorkflowTaskResponse
	client  workflowservice.WorkflowServiceClient
}

func newWorkflowWorker(
	env *testEnv,
	c *Cluster,
	tq *model.TaskQueue,
) *WorkflowWorker {
	return &WorkflowWorker{
		env:        env,
		ActorModel: stamp.NewActorModel(stamp.Trigger(c, trigger.CreateWorkflowWorker{TaskQueue: tq})),
		client:     c.tb.FrontendClient(),
	}
}

func (w *WorkflowWorker) OnTrigger(triggerID stamp.TriggerID, payload any) {
	switch t := payload.(type) {
	case proto.Message:
		_, _ = issueWorkflowRPC(w.env.context(), w.client, t, triggerID)
	default:
		panic(fmt.Sprintf("unhandled trigger %T", t))
	}
}

func (w *WorkflowWorker) SetHandler(fn func() trigger.WorkflowTaskResponse) {
	w.handler = fn
}

func (w *WorkflowWorker) CompleteTask(
	commands ...stamp.Gen[*commandpb.Command],
) trigger.CompleteWorkflowTask {
	return trigger.CompleteWorkflowTask{Commands: commands}
}
