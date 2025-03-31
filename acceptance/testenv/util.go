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
	"context"
	"reflect"
	"strings"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/server/common/headers"
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func issueWorkflowRPC(
	ctx context.Context,
	client workflowservice.WorkflowServiceClient,
	req proto.Message,
	triggerID stamp.TriggerID,
) (proto.Message, error) {
	// tagging the request with a trigger ID to match it to the action (see clusterMonitor)
	md := metadata.Pairs(triggerIdKey, string(triggerID))
	ctx, cancel := context.WithTimeout(headers.SetVersions(metadata.NewOutgoingContext(ctx, md)), 5*time.Second) // TODO: tweak this
	defer cancel()

	res := reflect.ValueOf(client).
		MethodByName(strings.TrimSuffix(string(req.ProtoReflect().Descriptor().Name()), "Request")).
		Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(req)})

	if res[0].IsNil() {
		return nil, res[1].Interface().(error)
	}
	return res[0].Interface().(proto.Message), nil
}
