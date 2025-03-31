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
	"fmt"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/server/acceptance/model"
	"go.temporal.io/server/acceptance/testenv/trigger"
	"go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/persistence/intercept"
	"go.temporal.io/server/common/softassert"
	"go.temporal.io/server/common/testing/stamp"
	"go.temporal.io/server/tests/testcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	triggerIdKey = "trigger-id"
)

type Cluster struct {
	stamp.ActorModel[*model.Cluster]
	env *testEnv
	tb  *testcore.FunctionalTestBase
}

func newCluster(env *testEnv, root *stamp.Root) *Cluster {
	res := &Cluster{
		env:        env,
		ActorModel: stamp.NewActorModel(stamp.Trigger(root, trigger.StartCluster{ClusterName: "local"})),
	}

	// TODO: add to FunctionalTestBase? measure!
	worker.SetBinaryChecksum("local")

	// TODO: cache and only start one at a time
	res.tb = &testcore.FunctionalTestBase{}

	res.tb.Logger = env.tl
	res.tb.SetT(env.t) // TODO: drop this; requires cluster setup to be decoupled from testify
	res.tb.SetupSuiteWithCluster(
		"../tests/testdata/es_cluster.yaml",
		// TODO: need to make this the 1st interceptor so it can observe every request
		testcore.WithAdditionalGrpcInterceptors(res.grpcInterceptor()),
		testcore.WithPersistenceInterceptor(res.dbInterceptor()))
	return res
}

func (c *Cluster) Stop() {
	c.tb.TearDownCluster()
}

func (c *Cluster) OnTrigger(triggerID stamp.TriggerID, payload any) {
	switch t := payload.(type) {
	case *persistence.CreateNamespaceRequest:
		// tagging the request with a trigger ID to match it to the action (see clusterMonitor)
		ctx := context.WithValue(c.env.context(), triggerIdKey, triggerID)
		_, _ = c.tb.GetTestCluster().TestBase().MetadataManager.CreateNamespace(ctx, t)

	case model.NewTaskQueue,
		model.NewWorkflowClient,
		model.NewWorkflowWorker:
		c.env.mdlEnv.HandleAction(triggerID, &model.Action{
			Cluster: model.ClusterName(c.GetID()),
			Request: t,
		})

	default:
		panic(fmt.Sprintf("unhandled trigger %T", t))
	}
}

func (c *Cluster) grpcInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		defer func() {
			if r := recover(); r != nil {
				softassert.Fail(c.env.tl, fmt.Sprintf("%v", r))
			}
		}()

		var triggerID stamp.TriggerID
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if _, ok := md[triggerIdKey]; ok {
				triggerID = stamp.TriggerID(md.Get(triggerIdKey)[0])
			}
		}

		act := &model.Action{
			Cluster:        model.ClusterName(c.GetID()),
			RequestID:      uuid.New(),
			RequestHeaders: md,
			Method:         info.FullMethod,
			Request:        req,
		}

		// handle request in model
		c.env.mdlEnv.HandleAction(triggerID, act)

		// TODO: support dropping the request before/after the response

		// process request
		resp, err := handler(ctx, req)

		// TODO: how to distinguish between expected and unexpected errors? error callback instead?
		//switch err.(type) {
		//case *serviceerror.InvalidArgument:
		//	c.env.Error(err.Error())
		//}

		// handle response in model
		if err == nil {
			act.Response = resp
		} else {
			act.ResponseErr = err
		}
		c.env.mdlEnv.HandleAction(triggerID, act)

		return resp, err
	}
}

func (c *Cluster) dbInterceptor() intercept.PersistenceInterceptor {
	return func(methodName string, fn func() (any, error), params ...any) error {
		defer func() {
			if r := recover(); r != nil {
				softassert.Fail(c.env.tl, fmt.Sprintf("%v", r))
			}
		}()

		var triggerID stamp.TriggerID
		var reqArgs []any
		for _, p := range params {
			if ctx, ok := p.(context.Context); ok {
				if ctxVal, ok := ctx.Value(triggerIdKey).(stamp.TriggerID); ok {
					triggerID = stamp.TriggerID(ctxVal)
				}
				continue // no value in adding context to the action
			}
			reqArgs = append(reqArgs, p)
		}

		act := &model.Action{
			Cluster:   model.ClusterName(c.GetID()),
			RequestID: uuid.New(),
			Method:    methodName,
		}
		if len(reqArgs) == 1 {
			act.Request = reqArgs[0]
		} else {
			act.Request = reqArgs
		}

		// handle request in model
		c.env.mdlEnv.HandleAction(triggerID, act)

		// process request
		resp, err := fn()

		// handle response in model
		act.Response = resp
		act.ResponseErr = err
		c.env.mdlEnv.HandleAction(triggerID, act)

		return err
	}
}
