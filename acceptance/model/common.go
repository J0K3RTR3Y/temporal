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

	"go.temporal.io/server/api/token/v1"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/persistence/serialization"
	"go.temporal.io/server/common/rpc/interceptor/logtags"
	"go.temporal.io/server/common/tasktoken"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	tokenSerializer = tasktoken.NewSerializer()
	blobSerializer  = serialization.NewSerializer()
	workflowTags    = logtags.NewWorkflowTags(tasktoken.NewSerializer(), log.NewTestLogger())
)

func mustDeserializeTaskToken(taskToken []byte) *token.Task {
	t, err := tokenSerializer.Deserialize(taskToken)
	if err != nil {
		panic(fmt.Errorf("failed to deserialize task token: %w", err))
	}
	return t
}

func findProtoValueByNameType[T any](
	msg proto.Message,
	name protoreflect.Name,
	kind protoreflect.Kind,
) T {
	return findProtoValue[T](msg, func(field protoreflect.FieldDescriptor, _ protoreflect.MessageDescriptor) bool {
		return field.Name() == name && field.Kind() == kind
	})
}

func findProtoValue[T any](
	msg proto.Message,
	matchFn func(protoreflect.FieldDescriptor, protoreflect.MessageDescriptor) bool,
) T {
	var res T
	if v := findProtoValueInternal(msg, nil, matchFn); v != nil {
		res = toValue[T](*v)
	}
	return res
}

// TODO: use expr lang for this instead?
func findProtoValueInternal(
	msg proto.Message,
	parent proto.Message,
	matchFn func(protoreflect.FieldDescriptor, protoreflect.MessageDescriptor) bool,
) *protoreflect.Value {
	var res *protoreflect.Value

	var parentDescr protoreflect.MessageDescriptor
	if parent != nil {
		parentDescr = parent.ProtoReflect().Descriptor()
	}

	msg.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if matchFn(fd, parentDescr) {
			res = &v
			return false
		}
		if fd.Kind() == protoreflect.MessageKind {
			if fd.IsList() {
				for i := 0; i < v.List().Len(); i++ {
					child := v.List().Get(i).Message().Interface()
					if res = findProtoValueInternal(child, msg, matchFn); res != nil {
						return false
					}
				}
			} else if !fd.IsMap() {
				child := v.Message().Interface()
				if res = findProtoValueInternal(child, msg, matchFn); res != nil {
					return false
				}
			}
		}
		return true
	})
	return res
}

func toValue[T any](v protoreflect.Value) T {
	var res T
	switch any(res).(type) {
	case string:
		return any(v.String()).(T)
	case int64:
		return any(v.Int()).(T)
	case float64:
		return any(v.Float()).(T)
	case bool:
		return any(v.Bool()).(T)
	case []uint8:
		return any(v.Bytes()).(T)
	default:
		panic(fmt.Sprintf("unsupported type %T", res))
	}
}
