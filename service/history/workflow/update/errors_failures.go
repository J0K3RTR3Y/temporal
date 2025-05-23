// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
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

package update

import (
	"errors"

	failurepb "go.temporal.io/api/failure/v1"
	"go.temporal.io/api/serviceerror"
)

var (
	registryClearedErr  = errors.New("update registry was cleared")
	workflowTaskFailErr = serviceerror.NewWorkflowNotReady("Unable to perform workflow execution update due to unexpected workflow task failure.")
)

var (
	unprocessedUpdateFailure = &failurepb.Failure{
		Message: "Workflow Update is rejected because it wasn't processed by worker. Probably, Workflow Update is not supported by the worker.",
		Source:  "Server",
		FailureInfo: &failurepb.Failure_ApplicationFailureInfo{ApplicationFailureInfo: &failurepb.ApplicationFailureInfo{
			Type:         "UnprocessedUpdate",
			NonRetryable: true,
		}},
	}

	acceptedUpdateCompletedWorkflowFailure = &failurepb.Failure{
		Message: "Workflow Update failed because the Workflow completed before the Update completed.",
		Source:  "Server",
		FailureInfo: &failurepb.Failure_ApplicationFailureInfo{ApplicationFailureInfo: &failurepb.ApplicationFailureInfo{
			Type:         "AcceptedUpdateCompletedWorkflow",
			NonRetryable: true,
		}},
	}
)
