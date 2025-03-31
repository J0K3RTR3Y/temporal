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
	"cmp"

	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	updatepb "go.temporal.io/api/update/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/acceptance/model"
	"go.temporal.io/server/common/testing/stamp"
)

type StartWorkflowUpdate struct {
	stamp.TriggerActor[*model.WorkflowClient]
	stamp.TriggerTarget[*model.WorkflowUpdate]
	Workflow       *model.Workflow `validate:"required"`
	UseLatestRunID stamp.Gen[bool]
	// TODO: change validator to flag non-pointer generators that are not provided
	WaitStage     stamp.Gen[enumspb.UpdateWorkflowExecutionLifecycleStage] `validate:"required"`
	UpdateID      stamp.GenProvider[stamp.ID]
	UpdateHandler stamp.GenProvider[stamp.ID]
	Payload       Payloads
	Header        Header
}

func (w StartWorkflowUpdate) Get(ctx stamp.GenContext) *workflowservice.UpdateWorkflowExecutionRequest {
	var runID string
	if !cmp.Or(w.UseLatestRunID, stamp.GenJust(true)).Get(ctx) {
		runID = w.Workflow.RunID().WaitGet(ctx)
	}

	return &workflowservice.UpdateWorkflowExecutionRequest{
		Namespace: string(w.Workflow.GetNamespace().GetID()),
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: string(w.Workflow.GetID()),
			RunId:      runID,
		},
		WaitPolicy: &updatepb.WaitPolicy{
			LifecycleStage: w.WaitStage.Get(ctx),
		},
		Request: &updatepb.Request{
			Meta: &updatepb.Meta{
				UpdateId: string(w.UpdateID.Get(ctx)),
				Identity: string(w.GetActor().GetID()),
			},
			Input: &updatepb.Input{
				Name:   string(w.UpdateHandler.Get(ctx)),
				Header: w.Header.Get(ctx),
				Args:   w.Payload.Get(ctx),
			},
		},
	}
}

type WorkflowUpdateRequest *updatepb.Request

type StartWorkflowUpdatePoll struct {
	stamp.TriggerActor[*model.WorkflowClient]
	stamp.TriggerTarget[*model.WorkflowUpdate]
	WorkflowUpdate *model.WorkflowUpdate
	WaitStage      stamp.Gen[enumspb.UpdateWorkflowExecutionLifecycleStage]
}

func (w StartWorkflowUpdatePoll) Get(ctx stamp.GenContext) *workflowservice.PollWorkflowExecutionUpdateRequest {
	workflow := w.WorkflowUpdate.GetWorkflow()
	return &workflowservice.PollWorkflowExecutionUpdateRequest{
		Namespace: string(workflow.GetID()),
		UpdateRef: &updatepb.UpdateRef{
			WorkflowExecution: &commonpb.WorkflowExecution{
				WorkflowId: string(workflow.GetID()),
				RunId:      workflow.RunID().WaitGet(ctx),
			},
			UpdateId: string(w.WorkflowUpdate.GetID()),
		},
		Identity: string(w.GetActor().GetID()),
		WaitPolicy: &updatepb.WaitPolicy{
			LifecycleStage: w.WaitStage.Get(ctx),
		},
	}
}

type StartUpdateWithStart struct {
	stamp.TriggerActor[*model.WorkflowClient]
	stamp.TriggerTarget[*model.WorkflowUpdate]
	StartWorkflowUpdate StartWorkflowUpdate
	StartWorkflow       StartWorkflow
	TaskQueue           *model.TaskQueue
}

func (w StartUpdateWithStart) Get(ctx stamp.GenContext) *workflowservice.ExecuteMultiOperationRequest {
	return &workflowservice.ExecuteMultiOperationRequest{
		Namespace: string(w.TaskQueue.GetNamespace().GetID()),
		//Operations: []*workflowservice.ExecuteMultiOperationRequest_Operation{
		//	{
		//
		//	},
		//}
	}
}

//	resp, err := c.workflowServiceClient.ExecuteMultiOperation(
//		NewContext(),
//		&workflowservice.ExecuteMultiOperationRequest{
//			Namespace: NamespaceName.Get(ns.Get(taskQueue.Get(wf))).String(),
//			Operations: []*workflowservice.ExecuteMultiOperationRequest_Operation{
//				{
//					Operation: &workflowservice.ExecuteMultiOperationRequest_Operation_StartWorkflow{
//						StartWorkflow: startWorkflowRequest(wf.ModelType),
//					},
//				},
//				{
//					Operation: &workflowservice.ExecuteMultiOperationRequest_Operation_UpdateWorkflow{
//						UpdateWorkflow: startUpdateRequest(upd.ModelType),
//					},
//				},
//			},
//		})
