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
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type (
	TaskQueue struct {
		stamp.Model[*TaskQueue, *Action]
		stamp.Scope[*Namespace]
	}
	NewTaskQueue struct {
		NamespaceName stamp.ID
		TaskQueueName stamp.ID
	}
)

func (t *TaskQueue) GetNamespace() *Namespace {
	return t.GetScope()
}

func (t *TaskQueue) Identity(action *Action) stamp.ID {
	switch t := action.Request.(type) {
	case NewTaskQueue:
		return t.TaskQueueName
	case NewWorkflowWorker:
		return t.TaskQueueName
	case NewWorkflowClient:
		return t.TaskQueueName
	case proto.Message:
		if v := findProtoValueInternal(t, nil, func(field protoreflect.FieldDescriptor, parent protoreflect.MessageDescriptor) bool {
			return parent != nil && parent.Name() == "TaskQueue" && field.Name() == "Name"
		}); v != nil {
			return stamp.ID(v.String())
		}
	}
	return ""
}

func (t *TaskQueue) OnAction(action *Action) stamp.Record {
	return nil
}
