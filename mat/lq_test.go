// Copyright ©2013 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat

import (
	"math/rand/v2"
	"testing"
)

func TestLQ(t *testing.T) {
	t.Parallel()
	const tol = 1e-14
	rnd := rand.New(rand.NewPCG(1, 1))
	for cas, test := range []struct {
		m, n int
	}{
		{5, 5},
		{5, 10},
	} {
		m := test.m
		n := test.n
		a := NewDense(m, n, nil)
		for i := 0; i < m; i++ {
			for j := 0; j < n; j++ {
				a.Set(i, j, rnd.NormFloat64())
			}
		}
		var want Dense
		want.CloneFrom(a)

		var lq LQ
		lq.Factorize(a)

		if !EqualApprox(a, &lq, tol) {
			t.Errorf("case %d: A and LQ are not equal", cas)
		}

		var l, q Dense
		lq.QTo(&q)

		if !isOrthonormal(&q, tol) {
			t.Errorf("Q is not orthonormal: m = %v, n = %v", m, n)
		}

		lq.LTo(&l)

		var got Dense
		got.Mul(&l, &q)
		if !EqualApprox(&got, &want, tol) {
			t.Errorf("LQ does not equal original matrix. \nWant: %v\nGot: %v", want, got)
		}
	}
}

func TestLQSolveTo(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, trans := range []bool{false, true} {
		for _, test := range []struct {
			m, n, bc int
		}{
			{5, 5, 1},
			{5, 10, 1},
			{5, 5, 3},
			{5, 10, 3},
		} {
			m := test.m
			n := test.n
			bc := test.bc
			a := NewDense(m, n, nil)
			for i := 0; i < m; i++ {
				for j := 0; j < n; j++ {
					a.Set(i, j, rnd.Float64())
				}
			}
			br := m
			if trans {
				br = n
			}
			b := NewDense(br, bc, nil)
			for i := 0; i < br; i++ {
				for j := 0; j < bc; j++ {
					b.Set(i, j, rnd.Float64())
				}
			}
			var x Dense
			lq := &LQ{}
			lq.Factorize(a)
			err := lq.SolveTo(&x, trans, b)
			if err != nil {
				t.Errorf("unexpected error from LQ solve: %v", err)
			}

			// Test that the normal equations hold.
			// Aᵀ * A * x = Aᵀ * b if !trans
			// A * Aᵀ * x = A * b if trans
			var lhs Dense
			var rhs Dense
			if trans {
				var tmp Dense
				tmp.Mul(a, a.T())
				lhs.Mul(&tmp, &x)
				rhs.Mul(a, b)
			} else {
				var tmp Dense
				tmp.Mul(a.T(), a)
				lhs.Mul(&tmp, &x)
				rhs.Mul(a.T(), b)
			}
			if !EqualApprox(&lhs, &rhs, 1e-10) {
				t.Errorf("Normal equations do not hold.\nLHS: %v\n, RHS: %v\n", lhs, rhs)
			}
		}
	}
	// TODO(btracey): Add in testOneInput when it exists.
}

func TestLQSolveToVec(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, trans := range []bool{false, true} {
		for _, test := range []struct {
			m, n int
		}{
			{5, 5},
			{5, 10},
		} {
			m := test.m
			n := test.n
			a := NewDense(m, n, nil)
			for i := 0; i < m; i++ {
				for j := 0; j < n; j++ {
					a.Set(i, j, rnd.Float64())
				}
			}
			br := m
			if trans {
				br = n
			}
			b := NewVecDense(br, nil)
			for i := 0; i < br; i++ {
				b.SetVec(i, rnd.Float64())
			}
			var x VecDense
			lq := &LQ{}
			lq.Factorize(a)
			err := lq.SolveVecTo(&x, trans, b)
			if err != nil {
				t.Errorf("unexpected error from LQ solve: %v", err)
			}

			// Test that the normal equations hold.
			// Aᵀ * A * x = Aᵀ * b if !trans
			// A * Aᵀ * x = A * b if trans
			var lhs Dense
			var rhs Dense
			if trans {
				var tmp Dense
				tmp.Mul(a, a.T())
				lhs.Mul(&tmp, &x)
				rhs.Mul(a, b)
			} else {
				var tmp Dense
				tmp.Mul(a.T(), a)
				lhs.Mul(&tmp, &x)
				rhs.Mul(a.T(), b)
			}
			if !EqualApprox(&lhs, &rhs, 1e-10) {
				t.Errorf("Normal equations do not hold.\nLHS: %v\n, RHS: %v\n", lhs, rhs)
			}
		}
	}
	// TODO(btracey): Add in testOneInput when it exists.
}

func TestLQSolveToCond(t *testing.T) {
	t.Parallel()
	for _, test := range []*Dense{
		NewDense(2, 2, []float64{1, 0, 0, 1e-20}),
		NewDense(2, 3, []float64{1, 0, 0, 0, 1e-20, 0}),
	} {
		m, _ := test.Dims()
		var lq LQ
		lq.Factorize(test)
		b := NewDense(m, 2, nil)
		var x Dense
		if err := lq.SolveTo(&x, false, b); err == nil {
			t.Error("No error for near-singular matrix in matrix solve.")
		}

		bvec := NewVecDense(m, nil)
		var xvec VecDense
		if err := lq.SolveVecTo(&xvec, false, bvec); err == nil {
			t.Error("No error for near-singular matrix in matrix solve.")
		}
	}
}
