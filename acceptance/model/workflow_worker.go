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

package model

import (
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/testing/stamp"
)

type (
	WorkflowWorker struct {
		stamp.Model[*WorkflowWorker, *Action]
		stamp.Scope[*TaskQueue]
	}
	NewWorkflowWorker struct {
		TaskQueueName stamp.ID
		WorkerName    stamp.ID
	}
)

func (w *WorkflowWorker) Identity(action *Action) stamp.ID {
	switch t := action.Request.(type) {
	case NewWorkflowWorker:
		return t.WorkerName
	case *workflowservice.PollWorkflowTaskQueueRequest:
		return stamp.ID(t.Identity)
	case *workflowservice.RespondWorkflowTaskCompletedRequest:
		return stamp.ID(t.Identity)
	case *workflowservice.RespondWorkflowTaskFailedRequest:
		return stamp.ID(t.Identity)
	}
	return ""
}

func (w *WorkflowWorker) OnAction(action *Action) stamp.Record {
	return nil
}
