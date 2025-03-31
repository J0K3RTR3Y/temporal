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
	"testing"

	"go.temporal.io/server/acceptance/model"
	"go.temporal.io/server/acceptance/testenv/trigger"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/testing/stamp"
	"go.temporal.io/server/common/testing/testlogger"
)

type (
	testEnv struct {
		t      *testing.T
		tl     *testlogger.TestLogger
		mdlEnv *stamp.ModelEnv[*model.Action]
	}
)

func NewEnv(t *testing.T) *testEnv {
	tl := testlogger.NewTestLogger(t, testlogger.FailOnExpectedErrorOnly)
	testlogger.DontPanicOnError(tl)
	tl.Expect(testlogger.Error, ".*", tag.FailedAssertion)

	mdlEnv := stamp.NewModelEnv[*model.Action](tl, *model.TemporalModel)
	mdlEnv.Root().SetTriggerHandler(
		func(triggerID stamp.TriggerID, payload any) {
			switch trg := payload.(type) {
			case trigger.StartCluster:
				mdlEnv.HandleAction(triggerID, &model.Action{Cluster: trg.ClusterName, Request: payload})
			default:
				t.Fatalf("unexpected trigger %T", payload)
			}
		})

	return &testEnv{
		t:      t,
		tl:     tl,
		mdlEnv: mdlEnv,
	}
}

func (e *testEnv) NewCluster() *Cluster {
	cluster := newCluster(e, e.mdlEnv.Root())
	e.t.Cleanup(func() {
		if e.tl.ResetFailureStatus() {
			panic(`Failing test as unexpected error logs were found.
				Look for 'Unexpected Error log encountered'.`)
		}
		cluster.Stop()
	})
	return cluster
}

func (e *testEnv) NewWorkflowClient(c *Cluster, tq *model.TaskQueue) *WorkflowClient {
	return newWorkflowClient(e, c, tq)
}

func (e *testEnv) NewWorkflowWorker(c *Cluster, tq *model.TaskQueue) *WorkflowWorker {
	return newWorkflowWorker(e, c, tq)
}

func (e *testEnv) SetupSingleTaskQueue() (*Cluster, *model.Namespace, *model.TaskQueue, *WorkflowClient, *WorkflowWorker) {
	cluster := e.NewCluster()
	ns := stamp.Trigger(cluster, trigger.CreateNamespace{})
	tq := stamp.Trigger(cluster, trigger.CreateTaskQueue{Namespace: ns})
	client := e.NewWorkflowClient(cluster, tq)
	worker := e.NewWorkflowWorker(cluster, tq)
	return cluster, ns, tq, client, worker
}

func (e *testEnv) context() context.Context {
	return context.Background()
}
