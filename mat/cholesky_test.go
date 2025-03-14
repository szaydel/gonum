// Copyright ©2013 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mat

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"testing"

	"gonum.org/v1/gonum/floats/scalar"
)

func TestCholesky(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		a *SymDense

		cond   float64
		want   *TriDense
		posdef bool
	}{
		{
			a: NewSymDense(3, []float64{
				4, 1, 1,
				0, 2, 3,
				0, 0, 6,
			}),
			cond: 37,
			want: NewTriDense(3, true, []float64{
				2, 0.5, 0.5,
				0, 1.3228756555322954, 2.0788046015507495,
				0, 0, 1.195228609334394,
			}),
			posdef: true,
		},
	} {
		_, n := test.a.Dims()
		for _, chol := range []*Cholesky{
			{},
			{chol: NewTriDense(n-1, true, nil)},
			{chol: NewTriDense(n, true, nil)},
			{chol: NewTriDense(n+1, true, nil)},
		} {
			ok := chol.Factorize(test.a)
			if ok != test.posdef {
				t.Errorf("unexpected return from Cholesky factorization: got: ok=%t want: ok=%t", ok, test.posdef)
			}
			fc := DenseCopyOf(chol.chol)
			if !Equal(fc, test.want) {
				t.Error("incorrect Cholesky factorization")
			}
			if math.Abs(test.cond-chol.cond) > 1e-13 {
				t.Errorf("Condition number mismatch: Want %v, got %v", test.cond, chol.cond)
			}
			var U TriDense
			chol.UTo(&U)
			aCopy := DenseCopyOf(test.a)
			var a Dense
			a.Mul(U.TTri(), &U)
			if !EqualApprox(&a, aCopy, 1e-14) {
				t.Error("unexpected Cholesky factor product")
			}
			var L TriDense
			chol.LTo(&L)
			a.Mul(&L, L.TTri())
			if !EqualApprox(&a, aCopy, 1e-14) {
				t.Error("unexpected Cholesky factor product")
			}
		}
	}
}

func TestCholeskyAt(t *testing.T) {
	t.Parallel()
	for _, test := range []*SymDense{
		NewSymDense(3, []float64{
			53, 59, 37,
			59, 83, 71,
			37, 71, 101,
		}),
	} {
		var chol Cholesky
		ok := chol.Factorize(test)
		if !ok {
			t.Fatalf("Matrix not positive definite")
		}
		n := test.SymmetricDim()
		cn := chol.SymmetricDim()
		if cn != n {
			t.Errorf("Cholesky size does not match. Got %d, want %d", cn, n)
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				got := chol.At(i, j)
				want := test.At(i, j)
				if math.Abs(got-want) > 1e-12 {
					t.Errorf("Cholesky at does not match at %d, %d. Got %v, want %v", i, j, got, want)
				}
			}
		}
	}
}

func TestCholeskySolveTo(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		a   *SymDense
		b   *Dense
		ans *Dense
	}{
		{
			a: NewSymDense(2, []float64{
				1, 0,
				0, 1,
			}),
			b:   NewDense(2, 1, []float64{5, 6}),
			ans: NewDense(2, 1, []float64{5, 6}),
		},
		{
			a: NewSymDense(3, []float64{
				53, 59, 37,
				0, 83, 71,
				37, 71, 101,
			}),
			b:   NewDense(3, 1, []float64{5, 6, 7}),
			ans: NewDense(3, 1, []float64{0.20745069393718094, -0.17421475529583694, 0.11577794010226464}),
		},
	} {
		var chol Cholesky
		ok := chol.Factorize(test.a)
		if !ok {
			t.Fatal("unexpected Cholesky factorization failure: not positive definite")
		}

		var x Dense
		err := chol.SolveTo(&x, test.b)
		if err != nil {
			t.Errorf("unexpected error from Cholesky solve: %v", err)
		}
		if !EqualApprox(&x, test.ans, 1e-12) {
			t.Error("incorrect Cholesky solve solution")
		}

		var ans Dense
		ans.Mul(test.a, &x)
		if !EqualApprox(&ans, test.b, 1e-12) {
			t.Error("incorrect Cholesky solve solution product")
		}
	}
}

func TestCholeskySolveCholTo(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		a, b *SymDense
	}{
		{
			a: NewSymDense(2, []float64{
				1, 0,
				0, 1,
			}),
			b: NewSymDense(2, []float64{
				1, 0,
				0, 1,
			}),
		},
		{
			a: NewSymDense(2, []float64{
				1, 0,
				0, 1,
			}),
			b: NewSymDense(2, []float64{
				2, 0,
				0, 2,
			}),
		},
		{
			a: NewSymDense(3, []float64{
				53, 59, 37,
				59, 83, 71,
				37, 71, 101,
			}),
			b: NewSymDense(3, []float64{
				2, -1, 0,
				-1, 2, -1,
				0, -1, 2,
			}),
		},
	} {
		var chola, cholb Cholesky
		ok := chola.Factorize(test.a)
		if !ok {
			t.Fatal("unexpected Cholesky factorization failure for a: not positive definite")
		}
		ok = cholb.Factorize(test.b)
		if !ok {
			t.Fatal("unexpected Cholesky factorization failure for b: not positive definite")
		}

		var x Dense
		err := chola.SolveCholTo(&x, &cholb)
		if err != nil {
			t.Errorf("unexpected error from Cholesky solve: %v", err)
		}

		var ans Dense
		ans.Mul(test.a, &x)
		if !EqualApprox(&ans, test.b, 1e-12) {
			var y Dense
			err := y.Solve(test.a, test.b)
			if err != nil {
				t.Errorf("unexpected error from dense solve: %v", err)
			}
			t.Errorf("incorrect Cholesky solve solution product\ngot solution:\n%.4v\nwant solution\n%.4v",
				Formatted(&x), Formatted(&y))
		}
	}
}

func TestCholeskySolveVecTo(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		a   *SymDense
		b   *VecDense
		ans *VecDense
	}{
		{
			a: NewSymDense(2, []float64{
				1, 0,
				0, 1,
			}),
			b:   NewVecDense(2, []float64{5, 6}),
			ans: NewVecDense(2, []float64{5, 6}),
		},
		{
			a: NewSymDense(3, []float64{
				53, 59, 37,
				0, 83, 71,
				0, 0, 101,
			}),
			b:   NewVecDense(3, []float64{5, 6, 7}),
			ans: NewVecDense(3, []float64{0.20745069393718094, -0.17421475529583694, 0.11577794010226464}),
		},
	} {
		var chol Cholesky
		ok := chol.Factorize(test.a)
		if !ok {
			t.Fatal("unexpected Cholesky factorization failure: not positive definite")
		}

		var x VecDense
		err := chol.SolveVecTo(&x, test.b)
		if err != nil {
			t.Errorf("unexpected error from Cholesky solve: %v", err)
		}
		if !EqualApprox(&x, test.ans, 1e-12) {
			t.Error("incorrect Cholesky solve solution")
		}

		var ans VecDense
		ans.MulVec(test.a, &x)
		if !EqualApprox(&ans, test.b, 1e-12) {
			t.Error("incorrect Cholesky solve solution product")
		}
	}
}

func TestCholeskyToSym(t *testing.T) {
	t.Parallel()
	for _, test := range []*SymDense{
		NewSymDense(3, []float64{
			53, 59, 37,
			0, 83, 71,
			0, 0, 101,
		}),
	} {
		var chol Cholesky
		ok := chol.Factorize(test)
		if !ok {
			t.Fatal("unexpected Cholesky factorization failure: not positive definite")
		}
		var s SymDense
		chol.ToSym(&s)

		if !EqualApprox(&s, test, 1e-12) {
			t.Errorf("Cholesky reconstruction not equal to original matrix.\nWant:\n% v\nGot:\n% v\n", Formatted(test), Formatted(&s))
		}
	}
}

func TestCloneCholesky(t *testing.T) {
	t.Parallel()
	for _, test := range []*SymDense{
		NewSymDense(3, []float64{
			53, 59, 37,
			0, 83, 71,
			0, 0, 101,
		}),
	} {
		var chol Cholesky
		ok := chol.Factorize(test)
		if !ok {
			panic("bad test")
		}
		var chol2 Cholesky
		chol2.Clone(&chol)

		if chol.cond != chol2.cond {
			t.Errorf("condition number mismatch from empty")
		}
		if !Equal(chol.chol, chol2.chol) {
			t.Errorf("chol mismatch from empty")
		}

		// Corrupt chol2 and try again
		chol2.cond = math.NaN()
		chol2.chol = NewTriDense(2, Upper, nil)
		chol2.Clone(&chol)
		if chol.cond != chol2.cond {
			t.Errorf("condition number mismatch from non-empty")
		}
		if !Equal(chol.chol, chol2.chol) {
			t.Errorf("chol mismatch from non-empty")
		}
	}
}

func TestCholeskyInverseTo(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 3, 5, 9} {
		data := make([]float64, n*n)
		for i := range data {
			data[i] = rnd.NormFloat64()
		}
		var s SymDense
		s.SymOuterK(1, NewDense(n, n, data))

		var chol Cholesky
		ok := chol.Factorize(&s)
		if !ok {
			t.Errorf("Bad test, cholesky decomposition failed")
		}

		var sInv SymDense
		err := chol.InverseTo(&sInv)
		if err != nil {
			t.Errorf("unexpected error from Cholesky inverse: %v", err)
		}

		var ans Dense
		ans.Mul(&sInv, &s)
		if !equalApprox(eye(n), &ans, 1e-8, false) {
			var diff Dense
			diff.Sub(eye(n), &ans)
			t.Errorf("SymDense times Cholesky inverse not identity. Norm diff = %v", Norm(&diff, 2))
		}
	}
}

func TestCholeskySymRankOne(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 2, 3, 4, 5, 7, 10, 20, 50, 100} {
		for k := 0; k < 50; k++ {
			// Construct a random positive definite matrix.
			data := make([]float64, n*n)
			for i := range data {
				data[i] = rnd.NormFloat64()
			}
			var a SymDense
			a.SymOuterK(1, NewDense(n, n, data))

			// Construct random data for updating.
			xdata := make([]float64, n)
			for i := range xdata {
				xdata[i] = rnd.NormFloat64()
			}
			x := NewVecDense(n, xdata)
			alpha := rnd.NormFloat64()

			// Compute the updated matrix directly. If alpha > 0, there are no
			// issues. If alpha < 0, it could be that the final matrix is not
			// positive definite, so instead switch the two matrices.
			aUpdate := NewSymDense(n, nil)
			if alpha > 0 {
				aUpdate.SymRankOne(&a, alpha, x)
			} else {
				aUpdate.CopySym(&a)
				a.Reset()
				a.SymRankOne(aUpdate, -alpha, x)
			}

			// Compare the Cholesky decomposition computed with Cholesky.SymRankOne
			// with that computed from updating A directly.
			var chol Cholesky
			ok := chol.Factorize(&a)
			if !ok {
				t.Errorf("Bad random test, Cholesky factorization failed")
				continue
			}

			var cholUpdate Cholesky
			ok = cholUpdate.SymRankOne(&chol, alpha, x)
			if !ok {
				t.Errorf("n=%v, alpha=%v: unexpected failure", n, alpha)
				continue
			}

			var aCompare SymDense
			cholUpdate.ToSym(&aCompare)
			if !EqualApprox(&aCompare, aUpdate, 1e-13) {
				t.Errorf("n=%v, alpha=%v: mismatch between updated matrix and from Cholesky:\nupdated:\n%v\nfrom Cholesky:\n%v",
					n, alpha, Formatted(aUpdate), Formatted(&aCompare))
			}
		}
	}

	for i, test := range []struct {
		a     *SymDense
		alpha float64
		x     []float64

		wantOk bool
	}{
		{
			// Update (to positive definite matrix).
			a: NewSymDense(4, []float64{
				1, 1, 1, 1,
				0, 2, 3, 4,
				0, 0, 6, 10,
				0, 0, 0, 20,
			}),
			alpha:  1,
			x:      []float64{0, 0, 0, 1},
			wantOk: true,
		},
		{
			// Downdate to singular matrix.
			a: NewSymDense(4, []float64{
				1, 1, 1, 1,
				0, 2, 3, 4,
				0, 0, 6, 10,
				0, 0, 0, 20,
			}),
			alpha:  -1,
			x:      []float64{0, 0, 0, 1},
			wantOk: false,
		},
		{
			// Downdate to positive definite matrix.
			a: NewSymDense(4, []float64{
				1, 1, 1, 1,
				0, 2, 3, 4,
				0, 0, 6, 10,
				0, 0, 0, 20,
			}),
			alpha:  -0.5,
			x:      []float64{0, 0, 0, 1},
			wantOk: true,
		},
		{
			// Issue #453.
			a:      NewSymDense(1, []float64{1}),
			alpha:  -1,
			x:      []float64{0.25},
			wantOk: true,
		},
	} {
		var chol Cholesky
		ok := chol.Factorize(test.a)
		if !ok {
			t.Errorf("Case %v: bad test, Cholesky factorization failed", i)
			continue
		}

		x := NewVecDense(len(test.x), test.x)
		ok = chol.SymRankOne(&chol, test.alpha, x)
		if !ok {
			if test.wantOk {
				t.Errorf("Case %v: unexpected failure from SymRankOne", i)
			}
			continue
		}
		if ok && !test.wantOk {
			t.Errorf("Case %v: expected a failure from SymRankOne", i)
		}

		a := test.a
		a.SymRankOne(a, test.alpha, x)

		var achol SymDense
		chol.ToSym(&achol)
		if !EqualApprox(&achol, a, 1e-13) {
			t.Errorf("Case %v: mismatch between updated matrix and from Cholesky:\nupdated:\n%v\nfrom Cholesky:\n%v",
				i, Formatted(a), Formatted(&achol))
		}
	}
}

func TestCholeskyExtendVecSym(t *testing.T) {
	t.Parallel()
	for cas, test := range []struct {
		a *SymDense
	}{
		{
			a: NewSymDense(3, []float64{
				4, 1, 1,
				0, 2, 3,
				0, 0, 6,
			}),
		},
	} {
		n := test.a.SymmetricDim()
		as := test.a.sliceSym(0, n-1)

		// Compute the full factorization to use later (do the full factorization
		// first to ensure the matrix is positive definite).
		var cholFull Cholesky
		ok := cholFull.Factorize(test.a)
		if !ok {
			panic("mat: bad test, matrix not positive definite")
		}

		var chol Cholesky
		ok = chol.Factorize(as)
		if !ok {
			panic("mat: bad test, subset is not positive definite")
		}
		row := NewVecDense(n, nil)
		for i := 0; i < n; i++ {
			row.SetVec(i, test.a.At(n-1, i))
		}

		var cholNew Cholesky
		ok = cholNew.ExtendVecSym(&chol, row)
		if !ok {
			t.Errorf("cas %v: update not positive definite", cas)
		}
		var a SymDense
		cholNew.ToSym(&a)
		if !EqualApprox(&a, test.a, 1e-12) {
			t.Errorf("cas %v: mismatch", cas)
		}

		// test in-place
		ok = chol.ExtendVecSym(&chol, row)
		if !ok {
			t.Errorf("cas %v: in-place update not positive definite", cas)
		}
		if !equalChol(&chol, &cholNew) {
			t.Errorf("cas %v: Cholesky different in-place vs. new", cas)
		}

		// Test that the factorization is about right compared with the direct
		// full factorization. Use a high tolerance on the condition number
		// since the condition number with the updated rule is approximate.
		if !equalApproxChol(&chol, &cholFull, 1e-12, 0.3) {
			t.Errorf("cas %v: updated Cholesky does not match full", cas)
		}
	}
}

func TestCholeskyScale(t *testing.T) {
	t.Parallel()
	for cas, test := range []struct {
		a *SymDense
		f float64
	}{
		{
			a: NewSymDense(3, []float64{
				4, 1, 1,
				0, 2, 3,
				0, 0, 6,
			}),
			f: 0.5,
		},
	} {
		var chol Cholesky
		ok := chol.Factorize(test.a)
		if !ok {
			t.Errorf("Case %v: bad test, Cholesky factorization failed", cas)
			continue
		}

		// Compare the update to a new Cholesky to an update in-place.
		var cholUpdate Cholesky
		cholUpdate.Scale(test.f, &chol)
		chol.Scale(test.f, &chol)
		if !equalChol(&chol, &cholUpdate) {
			t.Errorf("Case %d: cholesky mismatch new receiver", cas)
		}
		var sym SymDense
		chol.ToSym(&sym)
		var comp SymDense
		comp.ScaleSym(test.f, test.a)
		if !EqualApprox(&comp, &sym, 1e-14) {
			t.Errorf("Case %d: cholesky reconstruction doesn't match scaled matrix", cas)
		}

		var cholTest Cholesky
		cholTest.Factorize(&comp)
		if !equalApproxChol(&cholTest, &chol, 1e-12, 1e-12) {
			t.Errorf("Case %d: cholesky mismatch with scaled matrix. %v, %v", cas, cholTest.cond, chol.cond)
		}
	}
}

// equalApproxChol checks that the two Cholesky decompositions are equal.
func equalChol(a, b *Cholesky) bool {
	return Equal(a.chol, b.chol) && a.cond == b.cond
}

// equalApproxChol checks that the two Cholesky decompositions are approximately
// the same with the given tolerance on equality for the Triangular component and
// condition.
func equalApproxChol(a, b *Cholesky, matTol, condTol float64) bool {
	if !EqualApprox(a.chol, b.chol, matTol) {
		return false
	}
	return scalar.EqualWithinAbsOrRel(a.cond, b.cond, condTol, condTol)
}

func BenchmarkCholeskyFactorize(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			rnd := rand.New(rand.NewPCG(1, 1))

			data := make([]float64, n*n)
			for i := range data {
				data[i] = rnd.NormFloat64()
			}
			var a SymDense
			a.SymOuterK(1, NewDense(n, n, data))

			var chol Cholesky
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ok := chol.Factorize(&a)
				if !ok {
					panic("not positive definite")
				}
			}
		})
	}
}

func BenchmarkCholeskyToSym(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			rnd := rand.New(rand.NewPCG(1, 1))

			data := make([]float64, n*n)
			for i := range data {
				data[i] = rnd.NormFloat64()
			}
			var a SymDense
			a.SymOuterK(1, NewDense(n, n, data))

			var chol Cholesky
			ok := chol.Factorize(&a)
			if !ok {
				panic("not positive definite")
			}

			dst := NewSymDense(n, nil)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				chol.ToSym(dst)
			}
		})
	}
}

func BenchmarkCholeskyInverseTo(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			rnd := rand.New(rand.NewPCG(1, 1))

			data := make([]float64, n*n)
			for i := range data {
				data[i] = rnd.NormFloat64()
			}
			var a SymDense
			a.SymOuterK(1, NewDense(n, n, data))

			var chol Cholesky
			ok := chol.Factorize(&a)
			if !ok {
				panic("not positive definite")
			}

			dst := NewSymDense(n, nil)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := chol.InverseTo(dst)
				if err != nil {
					b.Fatalf("unexpected error from Cholesky inverse: %v", err)
				}
			}
		})
	}
}

func TestBandCholeskySolveTo(t *testing.T) {
	t.Parallel()

	const (
		nrhs = 4
		tol  = 1e-14
	)
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 2, 3, 5, 10} {
		for _, k := range []int{0, 1, n / 2, n - 1} {
			k := min(k, n-1)

			a := NewSymBandDense(n, k, nil)
			for i := 0; i < n; i++ {
				a.SetSymBand(i, i, rnd.Float64()+float64(n))
				for j := i + 1; j < min(i+k+1, n); j++ {
					a.SetSymBand(i, j, rnd.Float64())
				}
			}

			want := NewDense(n, nrhs, nil)
			for i := 0; i < n; i++ {
				for j := 0; j < nrhs; j++ {
					want.Set(i, j, rnd.NormFloat64())
				}
			}
			var b Dense
			b.Mul(a, want)

			for _, typ := range []SymBanded{a, (*basicSymBanded)(a)} {
				name := fmt.Sprintf("Case n=%d,k=%d,type=%T,nrhs=%d", n, k, typ, nrhs)

				var chol BandCholesky
				ok := chol.Factorize(typ)
				if !ok {
					t.Fatalf("%v: Factorize failed", name)
				}

				var got Dense
				err := chol.SolveTo(&got, &b)
				if err != nil {
					t.Errorf("%v: unexpected error from SolveTo: %v", name, err)
					continue
				}

				var resid Dense
				resid.Sub(want, &got)
				diff := Norm(&resid, math.Inf(1))
				if diff > tol {
					t.Errorf("%v: unexpected solution; diff=%v", name, diff)
				}

				got.Copy(&b)
				err = chol.SolveTo(&got, &got)
				if err != nil {
					t.Errorf("%v: unexpected error from SolveTo when dst==b: %v", name, err)
					continue
				}

				resid.Sub(want, &got)
				diff = Norm(&resid, math.Inf(1))
				if diff > tol {
					t.Errorf("%v: unexpected solution when dst==b; diff=%v", name, diff)
				}
			}
		}
	}
}

func TestBandCholeskySolveVecTo(t *testing.T) {
	t.Parallel()

	const tol = 1e-14
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 2, 3, 5, 10} {
		for _, k := range []int{0, 1, n / 2, n - 1} {
			k := min(k, n-1)

			a := NewSymBandDense(n, k, nil)
			for i := 0; i < n; i++ {
				a.SetSymBand(i, i, rnd.Float64()+float64(n))
				for j := i + 1; j < min(i+k+1, n); j++ {
					a.SetSymBand(i, j, rnd.Float64())
				}
			}

			want := NewVecDense(n, nil)
			for i := 0; i < n; i++ {
				want.SetVec(i, rnd.NormFloat64())
			}
			var b VecDense
			b.MulVec(a, want)

			for _, typ := range []SymBanded{a, (*basicSymBanded)(a)} {
				name := fmt.Sprintf("Case n=%d,k=%d,type=%T", n, k, typ)

				var chol BandCholesky
				ok := chol.Factorize(typ)
				if !ok {
					t.Fatalf("%v: Factorize failed", name)
				}

				var got VecDense
				err := chol.SolveVecTo(&got, &b)
				if err != nil {
					t.Errorf("%v: unexpected error from SolveVecTo: %v", name, err)
					continue
				}

				var resid VecDense
				resid.SubVec(want, &got)
				diff := Norm(&resid, math.Inf(1))
				if diff > tol {
					t.Errorf("%v: unexpected solution; diff=%v", name, diff)
				}

				got.CopyVec(&b)
				err = chol.SolveVecTo(&got, &got)
				if err != nil {
					t.Errorf("%v: unexpected error from SolveVecTo when dst==b: %v", name, err)
					continue
				}

				resid.SubVec(want, &got)
				diff = Norm(&resid, math.Inf(1))
				if diff > tol {
					t.Errorf("%v: unexpected solution when dst==b; diff=%v", name, diff)
				}
			}
		}
	}
}

func TestBandCholeskyAt(t *testing.T) {
	t.Parallel()

	const tol = 1e-14
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 2, 3, 5, 10} {
		for _, k := range []int{0, 1, n / 2, n - 1} {
			k := min(k, n-1)
			name := fmt.Sprintf("Case n=%d,k=%d", n, k)

			a := NewSymBandDense(n, k, nil)
			for i := 0; i < n; i++ {
				a.SetSymBand(i, i, rnd.Float64()+float64(n))
				for j := i + 1; j < min(i+k+1, n); j++ {
					a.SetSymBand(i, j, rnd.Float64())
				}
			}

			var chol BandCholesky
			ok := chol.Factorize(a)
			if !ok {
				t.Fatalf("%v: Factorize failed", name)
			}

			resid := NewDense(n, n, nil)
			for i := 0; i < n; i++ {
				for j := 0; j < n; j++ {
					resid.Set(i, j, math.Abs(a.At(i, j)-chol.At(i, j)))
				}
			}
			diff := Norm(resid, math.Inf(1))
			if diff > tol {
				t.Errorf("%v: unexpected result; diff=%v, want<=%v", name, diff, tol)
			}
		}
	}
}

func TestBandCholeskyDet(t *testing.T) {
	t.Parallel()

	const tol = 1e-14
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 2, 3, 5, 10} {
		for _, k := range []int{0, 1, n / 2, n - 1} {
			k := min(k, n-1)
			name := fmt.Sprintf("Case n=%d,k=%d", n, k)

			a := NewSymBandDense(n, k, nil)
			aSym := NewSymDense(n, nil)
			for i := 0; i < n; i++ {
				aii := rnd.Float64() + float64(n)
				a.SetSymBand(i, i, aii)
				aSym.SetSym(i, i, aii)
				for j := i + 1; j < min(i+k+1, n); j++ {
					aij := rnd.Float64()
					a.SetSymBand(i, j, aij)
					aSym.SetSym(i, j, aij)
				}
			}

			var chol BandCholesky
			ok := chol.Factorize(a)
			if !ok {
				t.Fatalf("%v: Factorize failed", name)
			}

			var cholDense Cholesky
			ok = cholDense.Factorize(aSym)
			if !ok {
				t.Fatalf("%v: dense Factorize failed", name)
			}

			want := cholDense.Det()
			got := chol.Det()
			if !scalar.EqualWithinRel(got, want, tol) {
				t.Errorf("%v: unexpected result; got=%v, want=%v (diff=%v)", name, got, want, math.Abs(got-want))
			}
		}
	}
}

func TestPivotedCholesky(t *testing.T) {
	t.Parallel()

	const tol = 1e-14
	src := rand.NewPCG(1, 1)
	for _, n := range []int{1, 2, 3, 4, 5, 10} {
		for _, rank := range []int{int(0.3 * float64(n)), int(0.7 * float64(n)), n} {
			name := fmt.Sprintf("n=%d, rank=%d", n, rank)

			// Generate a random symmetric semi-definite matrix A with the given rank.
			a := NewSymDense(n, nil)
			for i := 0; i < rank; i++ {
				x := randVecDense(n, 1, 1, src)
				a.SymRankOne(a, 1, x)
			}

			// Compute the pivoted Cholesky factorization of A.
			var chol PivotedCholesky
			ok := chol.Factorize(a, -1)

			// Check that the ok return matches the rank of A.
			if !ok && rank == n {
				t.Errorf("%s: unexpected factorization failure with full rank", name)
			}
			if ok && rank != n {
				t.Errorf("%s: unexpected factorization success with deficit rank", name)
			}

			// Check that the computed rank matches the rank of A.
			if chol.Rank() != rank {
				t.Errorf("%s: unexpected computed rank, got %d", name, chol.Rank())
			}

			// Check the size.
			r, c := chol.Dims()
			if r != n || c != n {
				t.Errorf("%s: unexpected dims: r=%d, c=%d", name, r, c)
			}
			if chol.SymmetricDim() != n {
				t.Errorf("%s: unexpected symmetric dim: dim=%d", name, chol.SymmetricDim())
			}

			// Compute the norm of the difference |P*Uᵀ*U*Pᵀ - A| using At.
			diff := NewDense(n, n, nil)
			for i := 0; i < n; i++ {
				for j := 0; j < n; j++ {
					diff.Set(i, j, chol.At(i, j)-a.At(i, j))
				}
			}
			res := Norm(diff, 1)
			if res > tol {
				t.Errorf("%s: unexpected result using At (|P*Uᵀ*U*Pᵀ - A|=%v)\ndiff=%.4g", name, res, Formatted(diff, Prefix("     ")))
			}

			// Compute the norm of the difference |P*Uᵀ*U*Pᵀ - A| using ColumnPivots and UTo.
			var u TriDense
			chol.UTo(&u)
			diff.Product(u.T(), &u)                        // Compute Uᵀ*U.
			diff.PermuteCols(chol.ColumnPivots(nil), true) // Multiply by Pᵀ from the right (inverse because we multiply by the transpose).
			diff.PermuteRows(chol.ColumnPivots(nil), true) // Multiply by P from the left (inverse because we pass a column permutation).
			diff.Sub(diff, a)
			res = Norm(diff, 1)
			if res > tol {
				t.Errorf("%s: unexpected result using ColumnPivots and UTo (|P*Uᵀ*U*Pᵀ - A|=%v)\ndiff=%.4g", name, res, Formatted(diff, Prefix("     ")))
			}

			// Compute the norm of the difference |P*Uᵀ*U*Pᵀ - A| using ColumnPivots and RawU.
			rawU := chol.RawU()
			diff.Product(rawU.T(), rawU)
			diff.PermuteCols(chol.ColumnPivots(nil), true)
			diff.PermuteRows(chol.ColumnPivots(nil), true)
			diff.Sub(diff, a)
			res = Norm(diff, 1)
			if res > tol {
				t.Errorf("%s: unexpected result using ColumnPivots and RawU (|P*Uᵀ*U*Pᵀ - A|=%v)\ndiff=%.4g", name, res, Formatted(diff, Prefix("     ")))
			}
		}
	}
}

func TestPivotedCholeskySolveTo(t *testing.T) {
	t.Parallel()

	const (
		nrhs = 4
		tol  = 1e-14
	)
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 2, 3, 5, 10} {
		a := NewSymDense(n, nil)
		for i := 0; i < n; i++ {
			a.SetSym(i, i, rnd.Float64()+float64(n))
			for j := i + 1; j < n; j++ {
				a.SetSym(i, j, rnd.Float64())
			}
		}

		want := NewDense(n, nrhs, nil)
		for i := 0; i < n; i++ {
			for j := 0; j < nrhs; j++ {
				want.Set(i, j, rnd.NormFloat64())
			}
		}

		var b Dense
		b.Mul(a, want)

		for _, typ := range []Symmetric{a, asBasicSymmetric(a)} {
			name := fmt.Sprintf("Case n=%d,type=%T,nrhs=%d", n, typ, nrhs)

			var chol PivotedCholesky
			ok := chol.Factorize(typ, -1)
			if !ok {
				t.Fatalf("%v: matrix not positive definite", name)
			}

			var got Dense
			err := chol.SolveTo(&got, &b)
			if err != nil {
				t.Errorf("%v: unexpected error from SolveTo: %v", name, err)
				continue
			}

			var resid Dense
			resid.Sub(want, &got)
			diff := Norm(&resid, math.Inf(1))
			if diff > tol {
				t.Errorf("%v: unexpected solution; diff=%v", name, diff)
			}

			got.Copy(&b)
			err = chol.SolveTo(&got, &got)
			if err != nil {
				t.Errorf("%v: unexpected error from SolveTo when dst==b: %v", name, err)
				continue
			}

			resid.Sub(want, &got)
			diff = Norm(&resid, math.Inf(1))
			if diff > tol {
				t.Errorf("%v: unexpected solution when dst==b; diff=%v", name, diff)
			}
		}
	}
}

func TestPivotedCholeskySolveVecTo(t *testing.T) {
	t.Parallel()

	const tol = 1e-14
	rnd := rand.New(rand.NewPCG(1, 1))
	for _, n := range []int{1, 2, 3, 5, 10} {

		a := NewSymDense(n, nil)
		for i := 0; i < n; i++ {
			a.SetSym(i, i, rnd.Float64()+float64(n))
			for j := i + 1; j < n; j++ {
				a.SetSym(i, j, rnd.Float64())
			}
		}

		want := NewVecDense(n, nil)
		for i := 0; i < n; i++ {
			want.SetVec(i, rnd.NormFloat64())
		}
		var b VecDense
		b.MulVec(a, want)

		for _, typ := range []Symmetric{a, asBasicSymmetric(a)} {
			name := fmt.Sprintf("Case n=%d,type=%T", n, typ)

			var chol PivotedCholesky
			ok := chol.Factorize(typ, -1)
			if !ok {
				t.Fatalf("%v: matrix not positive definite", name)
			}

			var got VecDense
			err := chol.SolveVecTo(&got, &b)
			if err != nil {
				t.Errorf("%v: unexpected error from SolveVecTo: %v", name, err)
				continue
			}

			var resid VecDense
			resid.SubVec(want, &got)
			diff := Norm(&resid, math.Inf(1))
			if diff > tol {
				t.Errorf("%v: unexpected solution; diff=%v", name, diff)
			}

			got.CopyVec(&b)
			err = chol.SolveVecTo(&got, &got)
			if err != nil {
				t.Errorf("%v: unexpected error from SolveVecTo when dst==b: %v", name, err)
				continue
			}

			resid.SubVec(want, &got)
			diff = Norm(&resid, math.Inf(1))
			if diff > tol {
				t.Errorf("%v: unexpected solution when dst==b; diff=%v", name, diff)
			}
		}
	}
}
