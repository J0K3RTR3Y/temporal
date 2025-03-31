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
	"cmp"
	"time"

	"github.com/pborman/uuid"
	enumspb "go.temporal.io/api/enums/v1"
	namespacepb "go.temporal.io/api/namespace/v1"
	"go.temporal.io/server/acceptance/model"
	persistencespb "go.temporal.io/server/api/persistence/v1"
	"go.temporal.io/server/common/namespace"
	"go.temporal.io/server/common/persistence"
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/protobuf/types/known/durationpb"
)

type CreateNamespace struct {
	stamp.TriggerActor[*model.Cluster]
	stamp.TriggerTarget[*model.Namespace]
	Name               stamp.GenProvider[stamp.ID]
	ID                 stamp.GenProvider[stamp.ID]
	NamespaceRetention stamp.Gen[time.Duration]
}

func (c CreateNamespace) Get(ctx stamp.GenContext) *persistence.CreateNamespaceRequest {
	return &persistence.CreateNamespaceRequest{
		Namespace: &persistencespb.NamespaceDetail{
			Info: &persistencespb.NamespaceInfo{
				Id:          namespace.ID(uuid.New()).String(), // TODO use generator
				Name:        string(c.Name.Get(ctx)),
				State:       enumspb.NAMESPACE_STATE_REGISTERED,
				Description: "namespace for acceptance tests",
			},
			Config: &persistencespb.NamespaceConfig{
				Retention:   durationpb.New(cmp.Or(c.NamespaceRetention, stamp.GenJust(24*time.Hour)).Get(ctx)),
				BadBinaries: &namespacepb.BadBinaries{Binaries: map[string]*namespacepb.BadBinaryInfo{}},
			},
			ReplicationConfig: &persistencespb.NamespaceReplicationConfig{
				ActiveClusterName: string(c.GetActor().GetID()),
				Clusters: []string{
					string(c.GetActor().GetID()),
				},
			},
		},
	}
}
