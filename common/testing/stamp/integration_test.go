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

package stamp

type (
	Foo struct {
		Model[*Foo]
		Scope[*Root]
	}
	FooBar struct {
		Model[*FooBar]
		Scope[*Foo]
	}
	request[T any] struct {
		id   string
		data T
	}
	response[T any] struct {
		data T
	}
	LastReqData           string
	LastRespData          string
	CompletedConversation bool
)

//func TestSend(t *testing.T) {
//	env := NewModelEnv(t)
//	RegisterModel[Foo](env)
//	RegisterModel[FooBar](env)
//
//	report := env.Send(request[string]{id: "a", data: "hello"})
//	require.Equal(t, 2, report.Len())
//	require.ErrorContains(t, report, "check for [spade.LastRespData] failed: 'Should not be zero, but was '")
//	require.ErrorContains(t, report, "check for [spade.CompletedConversation] failed: 'Should be true'")
//
//	models := 0
//	env.walk(func(_, mw modelWrapper) bool {
//		switch m := mw.(type) {
//		case *Foo:
//			require.Equal(t, "a", string(m.getDomainID()))
//			require.Equal(t, "hello", string(ReadVar[LastReqData](m)))
//			require.False(t, bool(ReadVar[CompletedConversation](m)))
//			models++
//		case *FooBar:
//			require.Equal(t, "a", string(m.getDomainID()))
//			require.Equal(t, "hello", string(ReadVar[LastReqData](m)))
//			require.False(t, bool(ReadVar[CompletedConversation](m)))
//			models++
//		}
//		return true
//	})
//	require.Equal(t, 2, models)
//
//	report = env.Send(request[string]{id: "a", data: "hello"}) // should not change anything
//	require.Equal(t, 2, report.Len())
//
//	models = 0
//	env.walk(func(_, mw modelWrapper) bool {
//		switch m := mw.(type) {
//		case *Foo:
//			require.Equal(t, "a", string(m.getDomainID()))
//			require.Equal(t, "hello", string(ReadVar[LastReqData](m)))
//			require.False(t, bool(ReadVar[CompletedConversation](m)))
//			models++
//		case *FooBar:
//			require.Equal(t, "a", string(m.getDomainID()))
//			require.Equal(t, "hello", string(ReadVar[LastReqData](m)))
//			require.False(t, bool(ReadVar[CompletedConversation](m)))
//			models++
//		}
//		return true
//	})
//	require.Equal(t, 2, models)
//
//	report = env.Send(request[string]{id: "a", data: "hello"}, response[string]{data: "hola"})
//	require.Equal(t, 0, report.Len())
//
//	models = 0
//	env.walk(func(_, mw modelWrapper) bool {
//		switch m := mw.(type) {
//		case *Foo:
//			require.Equal(t, "a", string(m.getDomainID()))             // updated
//			require.Equal(t, "hello", string(ReadVar[LastReqData](m))) // same
//			require.Equal(t, "hola", string(ReadVar[LastRespData](m)))
//			require.True(t, bool(ReadVar[CompletedConversation](m)))
//			models++
//		case *FooBar:
//			require.Equal(t, "a", string(m.getDomainID()))             // updated
//			require.Equal(t, "hello", string(ReadVar[LastReqData](m))) // same
//			require.Equal(t, "hola", string(ReadVar[LastRespData](m)))
//			require.True(t, bool(ReadVar[CompletedConversation](m)))
//			models++
//		}
//		return true
//	})
//	require.Equal(t, 2, models)
//}

func (f *Foo) IdRequest(req request[string]) ID { // TODO: make this a var like all the others?
	return ID(req.id)
}

func (f *Foo) OnRequest(req request[string]) LastReqData {
	return LastReqData(req.data)
}

func (f *Foo) OnResponse(req request[string], resp response[string]) LastRespData {
	return LastRespData(resp.data)
}

func (f *Foo) GetCompletedConversation(reqData LastReqData, respData LastRespData) CompletedConversation {
	return reqData != "" && respData != ""
}

func (fb *FooBar) IdRequest(req request[string]) ID {
	return ID(req.id)
}

func (fb *FooBar) OnRequest(req request[string]) LastReqData {
	return LastReqData(req.data)
}

func (fb *FooBar) Verify() Prop[bool] {
	panic("implement me")
	//return All(
	//	rule.NotZero[LastReqData](),
	//	rule.NotZero[LastRespData](),
	//	rule.True[CompletedConversation](),
	//)
}

func (fb *FooBar) OnResponse(req request[string], resp response[string]) LastRespData {
	return LastRespData(resp.data)
}

func (fb *FooBar) GetCompletedConversation(reqData LastReqData, respData LastRespData) CompletedConversation {
	return reqData != "" && respData != ""
}
