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

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
)

var (
	_            prop = Prop[bool]{}
	propType          = reflect.TypeFor[prop]()
	PropViolated      = errors.New("violated")
)

type (
	Prop[T propTypes] struct {
		_        T // have to use the type parameter to enforce compile-time checking
		owner    modelWrapper
		name     string
		typeOf   reflect.Type
		evalExpr string
		evalFn   func([]Record) T
		examples []example
	}
	prop interface {
		Validate() error
		String() string
		setName(string)
		eval(*evalContext) (any, error)
	}
	propTypes interface {
		// when updating this, make sure to also update the `typeCheckOption` switch below
		bool | int | string | ID
	}
	example struct {
		outcome any
		evalCtx *evalContext
	}
	propErr struct {
		prop string
		err  error
	}
	propOption[T propTypes] func(*Prop[T])
)

func WithEventually() propOption[bool] {
	return func(prop *Prop[bool]) {
		// TODO
		// - need to keep track over a few seconds
		// - fail the model if it doesn't happen
		// - and wait at the end of the scenario until timeout
	}
}

func WithExample[T propTypes](outcome T, record Record, records ...Record) propOption[T] {
	return func(prop *Prop[T]) {
		prop.examples = append(prop.examples, example{
			outcome: outcome,
			evalCtx: newEvalContext(append([]Record{record}, records...)...),
		})
	}
}

// See https://expr-lang.org/docs/language-definition
func PropExpr[T propTypes](
	owner modelWrapper,
	expr string,
	opts ...propOption[T],
) Prop[T] {
	r := Prop[T]{owner: owner, evalExpr: expr, typeOf: reflect.TypeFor[T]()}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

func NewProp[T propTypes](
	owner modelWrapper,
	evalFn func([]Record) T,
	opts ...propOption[T],
) Prop[T] {
	r := Prop[T]{owner: owner, evalFn: evalFn, typeOf: reflect.TypeFor[T]()}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

func (p Prop[T]) setName(name string) {
	p.name = name
}

func (p Prop[T]) String() string {
	return cmp.Or(p.name, p.evalExpr)
}

func (p Prop[T]) Validate() error {
	var failed bool
	errs := make([]error, len(p.examples))
	output := make([]any, len(p.examples))
	for i, ex := range p.examples {
		res, err := p.eval(ex.evalCtx)
		if err != nil {
			failed = true
			errs[i] = err
			continue
		}
		if res != ex.outcome {
			failed = true
			continue
		}
		output[i] = res
	}
	if failed {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("examples failed:\n"))
		sb.WriteString("\n")
		sb.WriteString(underlineStr("Type:\n"))
		sb.WriteString(p.typeOf.String())
		sb.WriteString("\n\n")
		sb.WriteString(underlineStr("Expr:\n"))
		sb.WriteString(p.evalExpr)
		sb.WriteString("\n")
		for i, ex := range p.examples {
			sb.WriteString("\n")
			sb.WriteString(underlineStr(fmt.Sprintf("Example #%d:\n", i+1)))
			sb.WriteString(fmt.Sprintf("input: %v\n", ex.evalCtx))
			if errs[i] != nil {
				sb.WriteString(fmt.Sprintf("%s: %v\n", redStr("error"), errs[i]))
				continue
			}
			sb.WriteString(fmt.Sprintf("output: %v\n", output[i]))
			sb.WriteString(fmt.Sprintf("expected: %v\n", ex.outcome))
		}
		return fmt.Errorf(sb.String())
	}
	return nil
}

func (p Prop[T]) Get() (T, *evalContext, error) {
	evalCtx := p.owner.getEvalCtx()
	res, err := p.eval(evalCtx)
	if res == nil {
		var zero T
		return zero, nil, err
	}
	return res.(T), evalCtx, err
}

func (p Prop[T]) GetOrDefault() T {
	res, _, _ := p.Get()
	return res
}

func (p Prop[T]) WaitGet(genCtx GenContext) T {
	timeout := time.After(2 * time.Second)           // TODO: take from genCtx
	ticker := time.NewTicker(100 * time.Millisecond) // TODO: take from genCtx
	defer ticker.Stop()

	var lastErr error
	for {
		select {
		case <-ticker.C:
			res, _, err := p.Get()
			if err == nil {
				return res
			}
			lastErr = err
		case <-timeout:
			panic(fmt.Errorf("prop %q failed to eval after timeout: %w", p.String(), lastErr))
		}
	}
}

func (p Prop[T]) eval(evalCtx *evalContext) (any, error) {
	if p.evalFn != nil {
		return p.evalFn(evalCtx.Recorder.Get()), nil
	}

	opts := []expr.Option{expr.Patch(safeFieldAccessPatcher{})}

	// type checking
	var typeCheckOption expr.Option
	switch p.typeOf {
	case reflect.TypeFor[bool]():
		typeCheckOption = expr.AsBool()
	case reflect.TypeFor[int]():
		typeCheckOption = expr.AsInt()
	case reflect.TypeFor[string](),
		reflect.TypeFor[ID]():
		// TODO?
		//typeCheckOption = expr.AsKind(reflect.String)
	default:
		return nil, fmt.Errorf("unsupported prop type %q", p.typeOf.String())
	}
	if typeCheckOption != nil {
		opts = append(opts, typeCheckOption, expr.WarnOnAny())
	}

	// compile
	program, err := expr.Compile(p.evalExpr, opts...)
	if err != nil {
		return nil, &propErr{prop: p.String(), err: fmt.Errorf("compilation failed: %w", err)}
	}

	// evaluate
	params := map[string]any{"records": evalCtx.Recorder.Get()}
	output, err := expr.Run(program, params)
	if err != nil {
		return nil, &propErr{prop: p.String(), err: fmt.Errorf("evaluation failed: %w", err)}
	}

	return output, nil
}

func (e propErr) Error() string {
	return fmt.Errorf("prop %q %w", e.prop, e.err).Error()
}

type safeFieldAccessPatcher struct{}

func (p safeFieldAccessPatcher) Visit(node *ast.Node) {
	memberNode, ok := (*node).(*ast.MemberNode)
	if !ok {
		return
	}
	ast.Patch(node, &ast.MemberNode{
		Node: &ast.ConditionalNode{
			Cond: &ast.BinaryNode{
				Operator: "in",
				Left:     memberNode.Property,
				Right:    memberNode.Node,
			},
			Exp1: memberNode.Node,
			Exp2: &ast.MapNode{},
		},
		Property: memberNode.Property,
		Optional: memberNode.Optional,
		Method:   memberNode.Method,
	})
}
