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

package litmus

import (
	"testing"

	commandpb "go.temporal.io/api/command/v1"
	enumspb "go.temporal.io/api/enums/v1"
	. "go.temporal.io/server/acceptance/testenv"
	. "go.temporal.io/server/acceptance/testenv/trigger"
	. "go.temporal.io/server/common/testing/stamp"
)

func TestWorkflowUpdate(t *testing.T) {
	t.Parallel()
	env := NewEnv(t)

	DescribeScenario(t, "Complete", func(s *Scenario) {
		_, _, tq, client, worker := env.SetupSingleTaskQueue()
		wf := Trigger(client, StartWorkflow{TaskQueue: tq})
		wft := Trigger(worker, PollWorkflowTask{TaskQueue: tq})
		Trigger(worker, CompleteWorkflowTask{WorkflowTask: wft})

		upd := Trigger(client, StartWorkflowUpdate{
			Workflow:  wf,
			WaitStage: GenJust(enumspb.UPDATE_WORKFLOW_EXECUTION_LIFECYCLE_STAGE_COMPLETED)})
		Trigger(worker, CompleteWorkflowTask{
			WorkflowTask: wft,
			Commands:     []Gen[*commandpb.Command]{CompleteWorkflowCommand{}},
		})

		s.WaitUntil(upd.IsComplete)
	})
}
