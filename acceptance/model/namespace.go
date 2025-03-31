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

	"go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type (
	Namespace struct {
		stamp.Model[*Namespace, *Action]
		stamp.Scope[*Cluster]
	}
	NamespaceCreatedRecord struct {
		InternalID string `validate:"required"`
	}
)

func (n *Namespace) Identity(action *Action) stamp.ID {
	switch t := action.Request.(type) {
	case NewTaskQueue:
		return t.NamespaceName
	case NewWorkflowClient:
		return t.TaskQueueName
	case NewWorkflowWorker:
		return t.TaskQueueName

	// GRPC
	case proto.Message:
		// Special case for when there's only the namespace ID in the request/response.
		if reqInternalID := n.extractNamespaceID(t); reqInternalID != "" {
			if mdlInternalID := n.InternalID().GetOrDefault(); reqInternalID == mdlInternalID {
				return n.GetID()
			}
		}
		return stamp.ID(findProtoValueByNameType[string](t, "namespace", protoreflect.StringKind))

	// persistence
	case *persistence.CreateNamespaceRequest:
		return stamp.ID(t.Namespace.Info.Name)
	case *persistence.InternalCreateNamespaceRequest:
		// InternalCreateNamespaceRequest is a special case for the test setup
		// which creates the namespace directly; instead of through the API.
		return stamp.ID(t.Name)
	case *persistence.InternalUpdateWorkflowExecutionRequest:
		if mdlInternalID := n.InternalID().GetOrDefault(); t.UpdateWorkflowMutation.ExecutionInfo.NamespaceId == mdlInternalID {
			return n.GetID()
		}
	}

	return ""
}

func (n *Namespace) extractNamespaceID(msg proto.Message) string {
	id := findProtoValueByNameType[string](msg, "namespace_id", protoreflect.StringKind)
	if id == "" {
		taskToken := findProtoValueByNameType[[]byte](msg, "task_token", protoreflect.BytesKind)
		if taskToken != nil {
			return mustDeserializeTaskToken(taskToken).NamespaceId
		}
	}
	return id
}

func (n *Namespace) OnAction(action *Action) stamp.Record {
	return cmp.Or(
		onDbResp(action, func(req *persistence.CreateNamespaceRequest, _ any) stamp.Record {
			return NamespaceCreatedRecord{InternalID: req.Namespace.Info.Id}
		}),
		onDbResp(action, func(req *persistence.InternalCreateNamespaceRequest, _ any) stamp.Record {
			return NamespaceCreatedRecord{InternalID: req.ID}
		}),
	)
}

func (n *Namespace) InternalID() stamp.Prop[string] {
	return stamp.PropExpr[string](
		n,
		`records | find(.InternalID != "") | get("InternalID")`,
		stamp.WithExample("ID", NamespaceCreatedRecord{InternalID: "ID"}),
	)
}
