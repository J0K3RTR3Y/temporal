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
	"fmt"
	"unicode"

	"pgregory.net/rapid"
)

// TODO: use Go fuzzer instead of rapid? what's the shrinking like?
// checkout https://pkg.go.dev/github.com/AdaLogics/go-fuzz-headers#section-readme

type (
	Gen[T any] interface {
		Get(GenContext) T
	}
	gen[T any] struct {
		gen      *rapid.Generator[T]
		variants []T
	}
	GenProvider[T genProvider[T]] struct {
		cached    bool
		generated T
		customGen *gen[T]
	}
	genProvider[T any] interface {
		Gen() Gen[T]
	}
	GenContext interface {
		Seed() int
	}
)

func Just[T genProvider[T]](val T) GenProvider[T] {
	return GenProvider[T]{
		customGen: &gen[T]{gen: rapid.Just(val)},
	}
}

//func GenRange[T ~int](min, max int) iter.Seq[T] {
//	return func(yield func(int) bool) {
//		times := GenInt(min, max).Next(s)
//		for i := 0; i < times; i++ {
//			if !yield(i) {
//				return
//			}
//		}
//	}
//}

func GenJust[T any](val T) Gen[T] {
	return &gen[T]{gen: rapid.Just(val)}
}

func GenInt[T ~int](min, max int) Gen[T] {
	return customGen[T](func(t *rapid.T) T {
		return T(rapid.IntRange(min, max).Draw(t, ""))
	})
}

func GenBool[T ~bool]() Gen[T] {
	return customGen[T](func(t *rapid.T) T {
		return T(rapid.Bool().Draw(t, ""))
	})
}

func GenName[T ~string]() Gen[T] {
	return gen[T]{
		gen: rapid.Custom(func(t *rapid.T) T {
			return T(fmt.Sprintf("%v-%v",
				rapid.SampledFrom(adjectives).Draw(t, ""),
				rapid.SampledFrom(names).Draw(t, "")))
		})}
}

func GenID[T ~string]() Gen[T] {
	return customGen[T](func(t *rapid.T) T {
		return T(rapid.StringOfN(rapid.RuneFrom(nil, unicode.ASCII_Hex_Digit), 12, 12, -1).Draw(t, ""))
	})
}

func GenPick[T any](variants []T) Gen[T] {
	return gen[T]{
		gen:      rapid.SampledFrom(variants),
		variants: variants,
	}
}

func (g gen[T]) Get(s GenContext) T {
	return g.gen.Example(s.Seed())
}

func (g GenProvider[T]) Get(s GenContext) T {
	if !g.cached {
		if g.customGen != nil {
			// use custom generator, if available
			g.generated = g.customGen.Get(s)
		} else {
			// otherwise, use default generator
			// TODO: need to cache the generator itself, too
			g.generated = g.generated.Gen().Get(s)
		}
		g.cached = true
	}
	return g.generated
}

//func (g Gen[T]) Variants() []any {
//	var variants []any
//	for _, v := range g.variants {
//		variants = append(variants, v)
//	}
//	return variants
//}

func customGen[T any](fn func(t *rapid.T) T) Gen[T] {
	return gen[T]{gen: rapid.Custom(fn)}
}
