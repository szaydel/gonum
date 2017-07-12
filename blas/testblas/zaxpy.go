// Copyright ©2017 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testblas

import "testing"

type Zaxpyer interface {
	Zaxpy(n int, alpha complex128, x []complex128, incX int, y []complex128, incY int)
}

func ZaxpyTest(t *testing.T, impl Zaxpyer) {
	for tc, test := range []struct {
		n          int
		alpha      complex128
		incX, incY int
		x, y       []float64
	}{
		{
			n:     1,
			alpha: 2 + 3i,
		},
		{
			n: 2,
		},
	} {
	}
}
