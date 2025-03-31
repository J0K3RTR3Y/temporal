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

package trigger

import (
	"reflect"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/server/common/testing/stamp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Payloads struct{}

// TODO: make random
func (g Payloads) Get(ctx stamp.GenContext) *commonpb.Payloads {
	return &commonpb.Payloads{
		Payloads: []*commonpb.Payload{Payload{}.Get(ctx)},
	}
}

type Payload struct{}

// TODO: make random
func (g Payload) Get(ctx stamp.GenContext) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{
			"meta_data_key": []byte(`meta_data_val`),
		},
		Data: []byte(`test data`),
	}
}

type Header struct{}

// TODO: make random
func (g Header) Get(ctx stamp.GenContext) *commonpb.Header {
	return &commonpb.Header{
		Fields: map[string]*commonpb.Payload{
			"header_field": Payload{}.Get(ctx),
		},
	}
}

func unmarshalAny[T proto.Message](a *anypb.Any) T {
	pb := new(T)
	ppb := reflect.ValueOf(pb).Elem()
	pbNew := reflect.New(reflect.TypeOf(pb).Elem().Elem())
	ppb.Set(pbNew)
	err := a.UnmarshalTo(*pb)
	if err != nil {
		panic(err)
	}
	return *pb
}

func marshalAny(pb proto.Message) *anypb.Any {
	a, err := anypb.New(pb)
	if err != nil {
		panic(err)
	}
	return a
}
