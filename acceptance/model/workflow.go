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
	"time"

	commandpb "go.temporal.io/api/command/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	protocolpb "go.temporal.io/api/protocol/v1"
	. "go.temporal.io/api/workflowservice/v1"
	enumsspb "go.temporal.io/server/api/enums/v1"
	"go.temporal.io/server/common"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type (
	Workflow struct {
		stamp.Model[*Workflow, *Action]
		stamp.Scope[*Namespace]
		// TODO: stamp.Embed[*WorkflowTask] for embedding sub-modules that share the same ID
	}
	WorkflowRunStartRecord struct {
		RequestID string `validate:"required"`
	}
	WorkflowRunStartedRecord struct {
		RunID string `validate:"required"`
	}
	WorkflowRunPersistRecord struct {
		RunID   string `validate:"required"`
		Durable bool   `validate:"required"`
		State   string `validate:"required"`
	}
	WorkflowRunEventRecord struct {
		Type  enumspb.EventType `validate:"required"`
		Attrs any               `validate:"required"`
	}
	WorkflowWorkerReplyRecord struct {
		Commands []*commandpb.Command `validate:"required"`
		Messages []*protocolpb.Message
	}
)

func (w *Workflow) GetNamespace() *Namespace {
	return w.GetScope()
}

func (w *Workflow) Identity(action *Action) stamp.ID {
	switch t := action.Request.(type) {
	// GRPC
	case proto.Message:
		// search through request
		for _, logTag := range workflowTags.Extract(t, action.Method) {
			switch logTag.Key() {
			case tag.WorkflowIDKey:
				return stamp.ID(logTag.Value().(string))
			}
		}
		// search through response
		for _, logTag := range workflowTags.Extract(action.Response, action.Method) {
			switch logTag.Key() {
			case tag.WorkflowIDKey:
				return stamp.ID(logTag.Value().(string))
			}
		}

	// persistence
	case *persistence.InternalUpdateWorkflowExecutionRequest:
		return stamp.ID(t.UpdateWorkflowMutation.ExecutionInfo.WorkflowId)
	}
	return ""
}

func (w *Workflow) OnAction(action *Action) stamp.Record {
	return cmp.Or(
		// GRPC
		onRpcReq(action, func(req *StartWorkflowExecutionRequest) stamp.Record {
			return WorkflowRunStartRecord{RequestID: req.RequestId}
		}),
		onRpcResp(action, func(req *StartWorkflowExecutionRequest, resp *StartWorkflowExecutionResponse) stamp.Record {
			return WorkflowRunStartedRecord{RunID: resp.RunId}
		}),
		onRpcResp(action, func(req *RespondWorkflowTaskCompletedRequest, resp *RespondWorkflowTaskCompletedResponse) stamp.Record {
			if len(req.Commands) == 0 {
				return nil
			}
			return WorkflowWorkerReplyRecord{
				Commands: common.CloneProtos(req.Commands),
				Messages: common.CloneProtos(req.Messages),
			}
		}),

		// database
		onDbResp(action, func(req *persistence.InternalCreateWorkflowExecutionRequest, _ any) stamp.Record {
			var res = []stamp.Record{WorkflowRunPersistRecord{
				Durable: true,
				RunID:   req.NewWorkflowSnapshot.RunID,
				State:   req.NewWorkflowSnapshot.ExecutionState.State.String(),
			}}
			for _, nodes := range req.NewWorkflowNewEvents {
				events, err := blobSerializer.DeserializeEvents(nodes.Node.Events)
				if err != nil {
					panic(err)
				}
				for _, evt := range events {
					res = append(res, WorkflowRunEventRecord{
						Type:  evt.GetEventType(),
						Attrs: common.CloneProto(evt.Attributes.(proto.Message)),
					})
				}
			}
			return res
		}),
		onDbResp(action, func(req *persistence.InternalUpdateWorkflowExecutionRequest, _ any) stamp.Record {
			return WorkflowRunPersistRecord{
				Durable: true,
				RunID:   req.UpdateWorkflowMutation.RunID,
				State:   req.UpdateWorkflowMutation.ExecutionState.State.String(),
			}
		}),
	)
}

func (w *Workflow) WasCreated() stamp.Prop[bool] {
	return stamp.PropExpr[bool](
		w,
		`records | any(.RunID != "")`,
		stamp.WithExample(true, &WorkflowRunStartedRecord{RunID: "run-id"}),
		stamp.WithExample(false, &WorkflowRunStartedRecord{}),
	)
}

func (w *Workflow) IsDurable() stamp.Prop[bool] {
	return stamp.PropExpr[bool](
		w,
		`records | any(.Durable == true)`,
		stamp.WithExample(true, &WorkflowRunPersistRecord{Durable: true}),
		stamp.WithExample(false, nil),
	)
}

func (w *Workflow) IsComplete() stamp.Prop[bool] {
	return stamp.PropExpr[bool](
		w,
		`records | any(.State == "Completed")`,
		stamp.WithExample(true, &WorkflowRunPersistRecord{
			State: enumsspb.WORKFLOW_EXECUTION_STATE_COMPLETED.String(),
		}),
		stamp.WithExample(false, nil),
	)
}

func (w *Workflow) RunID() stamp.Prop[string] {
	return stamp.PropExpr[string](
		w,
		`records | findLast(.RunID != "") | get("RunID")`,
		stamp.WithExample("2nd",
			&WorkflowRunStartedRecord{RunID: "1st"},
			&WorkflowRunStartedRecord{RunID: "2nd"},
		),
	)
}

//func (w *Workflow) DefaultWorkflowTaskTimeout() stamp.Rule {
//	return stamp.RuleExpr(
//		w,
//		`records | find(.Attrs.WorkflowTaskTimeout.Seconds != 0) == 10`,
//		stamp.WithExample(true,
//			WorkflowRunEventRecord{
//				Type: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
//				Attrs: &historypb.WorkflowExecutionStartedEventAttributes{
//					WorkflowTaskTimeout: durationpb.New(10 * time.Second),
//				},
//			}),
//	)
//}

func (w *Workflow) DefaultWorkflowTaskTimeout2() stamp.Rule {
	return stamp.NewRule(
		w,
		func(records []stamp.Record) bool {
			for _, record := range records {
				if record, ok := record.(WorkflowRunEventRecord); ok {
					attributes := record.Attrs.(*historypb.WorkflowExecutionStartedEventAttributes)
					if attributes.WorkflowTaskTimeout != nil {
						return attributes.WorkflowTaskTimeout.AsDuration() == 10*time.Second
					}
					return false
				}
			}
			return true
		},
		stamp.WithExample(true,
			WorkflowRunEventRecord{
				Type: enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
				Attrs: &historypb.WorkflowExecutionStartedEventAttributes{
					WorkflowTaskTimeout: durationpb.New(10 * time.Second),
				},
			}),
	)
}
