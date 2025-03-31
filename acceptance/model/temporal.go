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
	"fmt"
	"strings"

	"go.temporal.io/api/serviceerror"
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

var TemporalModel *stamp.ModelSet

type (
	Action struct {
		Cluster        ClusterName
		Request        any
		RequestID      string
		RequestHeaders metadata.MD
		Method         string
		Response       any
		ResponseErr    error
	}
)

func init() {
	ms := stamp.NewModelSet()
	stamp.RegisterModel[*Cluster](ms)
	stamp.RegisterModel[*Namespace](ms)
	stamp.RegisterModel[*TaskQueue](ms)
	stamp.RegisterModel[*WorkflowClient](ms)
	stamp.RegisterModel[*WorkflowWorker](ms)
	stamp.RegisterModel[*Workflow](ms)
	stamp.RegisterModel[*WorkflowTask](ms)
	stamp.RegisterModel[*WorkflowUpdate](ms)

	TemporalModel = ms
}

func onRpcReq[RQ proto.Message](action *Action, fn func(RQ) stamp.Record) stamp.Record {
	req, ok := action.Request.(RQ)
	if !ok {
		return nil
	}
	return fn(req)
}

func onRpcResp[RQ, RS proto.Message](action *Action, fn func(RQ, RS) stamp.Record) stamp.Record {
	req, ok := action.Request.(RQ)
	if !ok {
		return nil
	}
	if action.Response == nil {
		return nil
	}
	return fn(req, action.Response.(RS))
}

func onRpcErr[RQ proto.Message](action *Action, fn func(RQ, serviceerror.ServiceError) stamp.Record) stamp.Record {
	req, ok := action.Request.(RQ)
	if !ok {
		return nil
	}
	if action.ResponseErr == nil {
		return nil
	}
	return fn(req, action.ResponseErr.(serviceerror.ServiceError))
}

func onDbReq[RQ any](action *Action, fn func(RQ) stamp.Record) stamp.Record {
	req, ok := action.Request.(RQ)
	if !ok {
		return nil
	}
	return fn(req)
}

func onDbResp[RQ, RS any](action *Action, fn func(RQ, RS) stamp.Record) stamp.Record {
	req, ok := action.Request.(RQ)
	if !ok {
		return nil
	}
	if action.ResponseErr != nil {
		return nil
	}
	if action.Response == nil { // not all DB requests have a response
		var zero RS
		return fn(req, zero)
	}
	return fn(req, action.Response.(RS))
}

func (a *Action) String() string {
	var sb strings.Builder
	sb.WriteString("Action[\n")
	if a.Cluster != "" {
		sb.WriteString(fmt.Sprintf("  Cluster: %s,\n", a.Cluster))
	}
	if a.Request != nil {
		sb.WriteString(fmt.Sprintf("  Request: %T,\n", a.Request))
	}
	if a.RequestHeaders != nil {
		// TODO?
	}
	if a.RequestID != "" {
		sb.WriteString(fmt.Sprintf("  RequestID: %s,\n", a.RequestID))
	}
	if a.Method != "" {
		sb.WriteString(fmt.Sprintf("  Method: %s,\n", a.Method))
	}
	if a.Response != nil {
		sb.WriteString(fmt.Sprintf("  Response: %T,\n", a.Response))
	}
	if a.ResponseErr != nil {
		sb.WriteString(fmt.Sprintf("  ResponseErr: %T\n", a.ResponseErr))
	}
	sb.WriteString("]")
	return sb.String()
}
