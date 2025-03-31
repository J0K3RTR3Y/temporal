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
	"fmt"

	commandpb "go.temporal.io/api/command/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	updatepb "go.temporal.io/api/update/v1"
	. "go.temporal.io/api/workflowservice/v1"
	enumsspb "go.temporal.io/server/api/enums/v1"
	"go.temporal.io/server/api/matchingservice/v1"
	"go.temporal.io/server/common/testing/stamp"
	"go.temporal.io/server/common/testing/testmarker"
	"google.golang.org/protobuf/proto"
)

var (
	WorkflowUpdateAccepted = testmarker.New(
		"WorkflowUpdate.Accepted",
		"Update was accepted by the update validator on the worker.")
	WorkflowUpdateRejected = testmarker.New(
		"WorkflowUpdate.Rejected",
		"Update was rejected by the update validator on the worker.")
	WorkflowUpdateCompletedWithAccepted = testmarker.New(
		"WorkflowUpdate.Completed-with-Accepted",
		"Update was accepted and completed by the update handler on the worker in a single WorkflowTask.")
	WorkflowUpdateCompletedAfterAccepted = testmarker.New(
		"WorkflowUpdate.Completed-after-Accepted",
		"Update was completed by the update handler on the worker in a separate WorkflowTask from the Acceptance.")
	WorkflowUpdateResurrectedFromAcceptance = testmarker.New(
		"WorkflowUpdate.Resurrected-from-Acceptance",
		"Update was resurrected from an Acceptance response from the worker after being lost.")
	WorkflowUpdateResurrectedFromRejection = testmarker.New(
		"WorkflowUpdate.Resurrected-from-Rejection",
		"Update was resurrected from a Rejection response from the worker after being lost.")
)

type (
	WorkflowUpdate struct {
		stamp.Model[*WorkflowUpdate, *Action]
		stamp.Scope[*Workflow]
	}

	WorkflowUpdateStartRecord struct {
		Error serviceerror.ServiceError
	}
	WorkflowUpdatePollRecord    struct{}
	WorkflowUpdateOutcomeRecord struct {
		Outcome *updatepb.Outcome
	}
	WorkflowUpdateReqRecord struct {
		Request *updatepb.Request
	}
	WorkflowUpdateRespRecord struct {
		Message proto.Message
		Attrs   *commandpb.ProtocolMessageCommandAttributes
	}
)

func (u *WorkflowUpdate) GetWorkflow() *Workflow {
	return u.GetScope()
}

func (u *WorkflowUpdate) Identity(action *Action) stamp.ID {
	switch t := action.Request.(type) {
	case *UpdateWorkflowExecutionRequest:
		return stamp.ID(t.Request.Meta.UpdateId)
	case *PollWorkflowExecutionUpdateRequest:
		return stamp.ID(t.UpdateRef.UpdateId)
	case *PollWorkflowTaskQueueResponse:
		for _, msg := range t.Messages {
			if string(u.GetID()) == msg.ProtocolInstanceId {
				return u.GetID()
			}
		}
	case *RespondWorkflowTaskCompletedRequest:
		for _, msg := range t.Messages {
			if string(u.GetID()) == msg.ProtocolInstanceId {
				return u.GetID()
			}
		}
	}
	return ""
}

func (u *WorkflowUpdate) OnAction(action *Action) stamp.Record {
	return cmp.Or(
		// Update Start
		onRpcReq(action, func(req *UpdateWorkflowExecutionRequest) stamp.Record {
			fmt.Println(u.GetID(), "WorkflowUpdate.Start")
			return WorkflowUpdateStartRecord{}
		}),
		onRpcResp(action, func(req *UpdateWorkflowExecutionRequest, resp *UpdateWorkflowExecutionResponse) stamp.Record {
			fmt.Println(u.GetID(), "WorkflowUpdate.Start.Resp")
			return WorkflowUpdateOutcomeRecord{
				Outcome: resp.Outcome,
			}
		}),
		onRpcErr(action, func(req *UpdateWorkflowExecutionRequest, err serviceerror.ServiceError) stamp.Record {
			fmt.Println(u.GetID(), "WorkflowUpdate.Start.Error")
			return WorkflowUpdateStartRecord{Error: err}
		}),

		// Update Poll
		onRpcReq(action, func(req *PollWorkflowExecutionUpdateRequest) stamp.Record {
			return nil // TODO
		}),
		onRpcResp(action, func(req *PollWorkflowExecutionUpdateRequest, resp *PollWorkflowExecutionUpdateResponse) stamp.Record {
			return nil // TODO
		}),
		onRpcErr(action, func(req *PollWorkflowExecutionUpdateRequest, err serviceerror.ServiceError) stamp.Record {
			return nil // TODO
		}),

		// Task Matching
		onRpcResp(action, func(req *matchingservice.AddWorkflowTaskRequest, resp *matchingservice.AddWorkflowTaskResponse) stamp.Record {
			return WorkflowTaskTransferredRecord{
				Speculative: len(action.RequestHeaders.Get("x-speculative-workflow-task")) > 0,
			}
		}),

		// Task Poll
		onRpcResp(action, func(req *PollWorkflowTaskQueueRequest, resp *PollWorkflowTaskQueueResponse) stamp.Record {
			var records []stamp.Record
			for _, msg := range resp.Messages {
				if string(u.GetID()) != msg.ProtocolInstanceId {
					continue
				}

				body, err := msg.Body.UnmarshalNew()
				if err != nil {
					// TODO: record error
					continue
				}

				switch updMsg := body.(type) {
				case *updatepb.Request:
					records = append(records, WorkflowUpdateReqRecord{Request: updMsg})
				default:
					// TODO: record error
				}
			}
			return records
		}),

		// Task Completion
		onRpcResp(action, func(req *RespondWorkflowTaskCompletedRequest, resp *RespondWorkflowTaskCompletedResponse) stamp.Record {
			var commandAttrs []*commandpb.ProtocolMessageCommandAttributes
			for _, cmd := range req.Commands {
				switch cmd.CommandType {
				case enumspb.COMMAND_TYPE_PROTOCOL_MESSAGE:
					commandAttrs = append(commandAttrs, cmd.GetProtocolMessageCommandAttributes())
				}
			}

			var records []stamp.Record
			for _, msg := range req.Messages {
				if string(u.GetID()) != msg.ProtocolInstanceId {
					continue
				}

				body, err := msg.Body.UnmarshalNew()
				if err != nil {
					// TODO: record error
					continue
				}

				switch updMsg := body.(type) {
				case *updatepb.Acceptance:
					record := WorkflowUpdateRespRecord{Message: updMsg}
					record.Attrs, commandAttrs = commandAttrs[0], commandAttrs[1:]
					records = append(records, record)
				case *updatepb.Rejection:
					records = append(records, WorkflowUpdateRespRecord{Message: updMsg})
					// NOTE: there is no command for a rejection
				case *updatepb.Response:
					record := WorkflowUpdateRespRecord{Message: updMsg}
					record.Attrs, commandAttrs = commandAttrs[0], commandAttrs[1:]
					records = append(records, record)
				default:
					panic(fmt.Sprintf("unknown message type: %T", body))
				}
			}
			return records
		}),
	)
}

// - verify speculative task queue was used *when expected* (check persistence writes ...)
// - only one roundtrip to worker if there are no faults
// - server receives correct worker outcome
// - update request message has correct eventId
// - ResetHistoryEventId is zero?
// - poll after completion returns same result (and UpdateRef contains correct RunID, too)
// - on completion: history contains correct events
// - on rejection: history is reset (or not if not possible!)
// - Update is correctly processed when worker *also* sends another command
// - Update is delivered as Normal WFT when it's on the WF's first one
// - when there's a buffered event (signal), then ...

// with Alex
// - when receiving a history from a Poll, the last event must be WFTStarted with
//   a scheduledEvent attribute pointing to the previous event, which is WFTScheduled
// - Update received by worker matches requested Update: Update ID, Update Handler etc.
// - eventually you either get an error or an Outcome
// - Update response from worker matches WorkflowWorker's response data: Update ID etc.
// - when you poll after completion, you get the same outcome
// - if Update is completed, the history will end with
//   WorkflowExecutionUpdateAccepted and WorkflowExecutionUpdateCompleted,
//   where `AcceptedEventId` points to WorkflowExecutionUpdateAccepted,
//   and `AcceptedRequestSequencingEventId` points at its `WorkflowTaskScheduled`
// - if a speculative Update completes successfully, the speculative WFT is part of the
//   history
// - (need different generators for "with RunID" and "without RunID"

func (u *WorkflowUpdate) IsComplete() stamp.Prop[bool] {
	return stamp.PropExpr[bool](
		u,
		`records | any(.State == "Completed")`,
		stamp.WithExample(true, &WorkflowRunPersistRecord{
			State: enumsspb.WORKFLOW_EXECUTION_STATE_COMPLETED.String(),
		}),
		stamp.WithExample(false, nil),
	)
}

//u.Assert.Equal(string(id), updMsg.AcceptedRequest.Meta.UpdateId)
//u.Assert.Equal(string(id), updMsg.AcceptedRequestMessageId)
// TODO: updMsg.AcceptedRequestSequencingEventId

//u.Assert.Equal(string(id), updMsg.RejectedRequest.Meta.UpdateId)
//u.Assert.Equal(string(id), updMsg.RejectedRequestMessageId)
//u.Assert.NotZero(updMsg.Failure)
//u.Assert.NotZero(updMsg.Failure.Message)
// TODO: updMsg.RejectedRequestSequencingEventId

//u.Assert.Equal(string(id), updMsg.Meta.UpdateId)
//u.Assert.NotNil(updMsg.Outcome)

// ==== triggers

// TODO: listen for parent events; such as state change to "dead"
