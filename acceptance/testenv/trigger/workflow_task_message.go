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
	failurepb "go.temporal.io/api/failure/v1"
	protocolpb "go.temporal.io/api/protocol/v1"
	updatepb "go.temporal.io/api/update/v1"
	"go.temporal.io/server/common/payloads"
	"go.temporal.io/server/common/testing/stamp"
)

type WorkflowUpdateAcceptMessage struct {
	Request *protocolpb.Message `validate:"required"`
}

func (m WorkflowUpdateAcceptMessage) Get(ctx stamp.GenContext) *protocolpb.Message {
	req := unmarshalAny[*updatepb.Request](m.Request.GetBody())
	return &protocolpb.Message{
		Id:                 "",
		SequencingId:       nil,
		ProtocolInstanceId: req.Meta.GetUpdateId(),
		Body: marshalAny(&updatepb.Acceptance{
			AcceptedRequestMessageId:         m.Request.GetId(),
			AcceptedRequestSequencingEventId: m.Request.GetEventId(),
			AcceptedRequest:                  req,
		}),
	}
}

type WorkflowUpdateRejectionMessage struct {
	Request *protocolpb.Message    `validate:"required"`
	Failure *WorkflowUpdateFailure `validate:"required"`
}

func (m WorkflowUpdateRejectionMessage) Get(ctx stamp.GenContext) *protocolpb.Message {
	req := unmarshalAny[*updatepb.Request](m.Request.GetBody())
	return &protocolpb.Message{
		Id:                 "",
		SequencingId:       nil,
		ProtocolInstanceId: req.Meta.GetUpdateId(),
		Body: marshalAny(&updatepb.Rejection{
			RejectedRequestMessageId:         m.Request.GetId(),
			RejectedRequestSequencingEventId: m.Request.GetEventId(),
			RejectedRequest:                  req,
			Failure:                          m.Failure.Get(ctx),
		}),
	}
}

type WorkflowUpdateCompletionMessage struct {
	Request *protocolpb.Message    `validate:"required"`
	Outcome *WorkflowUpdateOutcome `validate:"required"`
}

func (m WorkflowUpdateCompletionMessage) Get(ctx stamp.GenContext) *protocolpb.Message {
	req := unmarshalAny[*updatepb.Request](m.Request.GetBody())
	return &protocolpb.Message{
		Id:                 "",
		SequencingId:       nil,
		ProtocolInstanceId: req.Meta.GetUpdateId(),
		Body: marshalAny(&updatepb.Response{
			Meta:    req.Meta,
			Outcome: m.Outcome.Get(ctx),
		}),
	}
}

type WorkflowUpdateOutcome struct {
	// TODO
}

func (o WorkflowUpdateOutcome) Get(ctx stamp.GenContext) *updatepb.Outcome {
	return &updatepb.Outcome{
		Value: &updatepb.Outcome_Success{
			Success: payloads.EncodeString("success"),
		},
	}
}

type WorkflowUpdateFailure struct {
	// TODO
}

func (f WorkflowUpdateFailure) Get(ctx stamp.GenContext) *failurepb.Failure {
	return &failurepb.Failure{
		Message: "rejection",
		FailureInfo: &failurepb.Failure_ApplicationFailureInfo{
			ApplicationFailureInfo: &failurepb.ApplicationFailureInfo{},
		},
	}
}
