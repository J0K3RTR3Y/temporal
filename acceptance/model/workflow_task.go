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
	"cmp"
	"encoding/base64"

	. "go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/api/matchingservice/v1"
	"go.temporal.io/server/common/testing/stamp"
	"go.temporal.io/server/common/testing/testmarker"
)

var (
	SpeculativeWorkflowTaskLost = testmarker.New(
		"WorkflowTask.Speculative.Lost",
		"Speculative Workflow Task was lost.")
	SpeculativeWorkflowTaskStale = testmarker.New(
		"WorkflowTask.Speculative.Stale",
		"Speculative Workflow Task was stale.")
	SpeculativeWorkflowTaskConverted = testmarker.New(
		"WorkflowTask.Speculative.Converted",
		"Speculative Workflow Task was converted to a normal Workflow Task.")
)

type (
	WorkflowTask struct {
		stamp.Model[*WorkflowTask, *Action]
		stamp.Scope[*Workflow]
	}
	WorkflowTaskTransferredRecord struct {
		Speculative bool
	}
	WorkflowTaskPolledRecord struct {
		Token string
	}
)

func (w *WorkflowTask) GetWorkflow() *Workflow {
	return w.GetScope()
}

func (w *WorkflowTask) GetNamespace() *Namespace {
	return w.GetScope().GetNamespace()
}

func (w *WorkflowTask) Identity(_ *Action) stamp.ID {
	// Since WorkflowTask has a 1:1 relationship with a Workflow, we can use the Workflow's ID.
	return w.GetWorkflow().GetID()
}

func (w *WorkflowTask) OnAction(action *Action) stamp.Record {
	return cmp.Or(
		onRpcResp(action, func(req *PollWorkflowTaskQueueRequest, resp *PollWorkflowTaskQueueResponse) stamp.Record {
			return WorkflowTaskPolledRecord{
				Token: base64.StdEncoding.EncodeToString(resp.TaskToken),
			}
		}),
		onRpcResp(action, func(req *matchingservice.AddWorkflowTaskRequest, resp *matchingservice.AddWorkflowTaskResponse) stamp.Record {
			return WorkflowTaskTransferredRecord{
				Speculative: len(action.RequestHeaders.Get("x-speculative-workflow-task")) > 0,
			}
		}),
	)
}

func (w *WorkflowTask) WasPolled() stamp.Prop[bool] {
	return stamp.PropExpr[bool](
		w,
		`records | any(.Token != "")`,
		stamp.WithExample(true, &WorkflowTaskPolledRecord{Token: "token"}),
		stamp.WithExample(false, &WorkflowTaskPolledRecord{}),
	)
}

func (w *WorkflowTask) Token() stamp.Prop[string] {
	return stamp.PropExpr[string](
		w,
		`records | findLast(.Token != "") | get("Token")`,
		stamp.WithExample(
			"2nd",
			&WorkflowTaskPolledRecord{Token: "1st"},
			&WorkflowTaskPolledRecord{Token: "2nd"},
		),
	)
}
