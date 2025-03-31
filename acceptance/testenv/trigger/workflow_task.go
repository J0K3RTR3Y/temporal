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
// FITNESS FOR A PARTICULAR PURPOSE AND NGetONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package trigger

import (
	"encoding/base64"

	commandpb "go.temporal.io/api/command/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/acceptance/model"
	"go.temporal.io/server/common/testing/stamp"
)

type PollWorkflowTask struct {
	stamp.TriggerActor[*model.WorkflowWorker]
	stamp.TriggerTarget[*model.WorkflowTask]
	TaskQueue *model.TaskQueue
}

func (t PollWorkflowTask) Get(_ stamp.GenContext) *workflowservice.PollWorkflowTaskQueueRequest {
	return &workflowservice.PollWorkflowTaskQueueRequest{
		Namespace: string(t.TaskQueue.GetNamespace().GetID()),
		TaskQueue: &taskqueue.TaskQueue{
			Kind: enumspb.TASK_QUEUE_KIND_NORMAL,
			Name: string(t.TaskQueue.GetID()),
		},
		Identity: string(t.GetActor().GetID()),
	}
}

type WorkflowTaskResponse interface {
	Complete() *CompleteWorkflowTask
	Fail() *FailWorkflowTask
}

type CompleteWorkflowTask struct {
	stamp.TriggerActor[*model.WorkflowWorker]
	stamp.TriggerTarget[*model.WorkflowTask]
	WorkflowTask *model.WorkflowTask
	Commands     []stamp.Gen[*commandpb.Command]
}

func (t CompleteWorkflowTask) Get(ctx stamp.GenContext) *workflowservice.RespondWorkflowTaskCompletedRequest {
	tokenBytes, err := base64.StdEncoding.DecodeString(t.WorkflowTask.Token().WaitGet(ctx))
	if err != nil {
		panic(err)
	}

	var commands []*commandpb.Command
	for _, cmd := range t.Commands {
		commands = append(commands, cmd.Get(ctx))
	}

	return &workflowservice.RespondWorkflowTaskCompletedRequest{
		Identity:  string(t.GetActor().GetID()),
		Namespace: string(t.WorkflowTask.GetNamespace().GetID()),
		TaskToken: tokenBytes,
		Commands:  commands,
	}
}

func (t CompleteWorkflowTask) Complete() *CompleteWorkflowTask {
	return &t
}

func (t CompleteWorkflowTask) Fail() *FailWorkflowTask {
	return nil
}

type FailWorkflowTask struct {
	stamp.TriggerActor[*model.WorkflowWorker]
	stamp.TriggerTarget[*model.WorkflowTask]
	WorkflowTask *model.WorkflowTask
	// TODO
}

func (t FailWorkflowTask) Get(ctx stamp.GenContext) *workflowservice.RespondWorkflowTaskFailedRequest {
	return &workflowservice.RespondWorkflowTaskFailedRequest{
		// TODO
	}
}

func (t FailWorkflowTask) Complete() *CompleteWorkflowTask {
	return nil
}

func (t FailWorkflowTask) Fail() *FailWorkflowTask {
	return &t
}
