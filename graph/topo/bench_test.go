// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package topo

import (
	"fmt"
	"testing"

	"golang.org/x/exp/rand"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/graphs/gen"
	"gonum.org/v1/gonum/graph/simple"
)

var (
	gnpDirected_10_tenth   = gnpDirected(10, 0.1)
	gnpDirected_100_tenth  = gnpDirected(100, 0.1)
	gnpDirected_1000_tenth = gnpDirected(1000, 0.1)
	gnpDirected_10_half    = gnpDirected(10, 0.5)
	gnpDirected_100_half   = gnpDirected(100, 0.5)
	gnpDirected_1000_half  = gnpDirected(1000, 0.5)

	gnpUndirected_10_tenth  = gnpUndirected(10, 0.1)
	gnpUndirected_100_tenth = gnpUndirected(100, 0.1)
	gnpUndirected_10_half   = gnpUndirected(10, 0.5)
	gnpUndirected_100_half  = gnpUndirected(100, 0.5)
)

func gnpDirected(n int, p float64) graph.Directed {
	g := simple.NewDirectedGraph()
	err := gen.Gnp(g, n, p, rand.NewSource(1))
	if err != nil {
		panic(fmt.Sprintf("topo: bad test: %v", err))
	}
	return g
}

func gnpUndirected(n int, p float64) graph.Undirected {
	g := simple.NewUndirectedGraph()
	err := gen.Gnp(g, n, p, rand.NewSource(1))
	if err != nil {
		panic(fmt.Sprintf("topo: bad test: %v", err))
	}
	return g
}

func benchmarkTarjanSCC(b *testing.B, g graph.Directed) {
	var sccs [][]graph.Node
	for i := 0; i < b.N; i++ {
		sccs = TarjanSCC(g)
	}
	if len(sccs) == 0 {
		b.Fatal("unexpected number zero-sized SCC set")
	}
}

func BenchmarkTarjanSCCGnp_10_tenth(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_10_tenth)
}
func BenchmarkTarjanSCCGnp_100_tenth(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_100_tenth)
}
func BenchmarkTarjanSCCGnp_1000_tenth(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_1000_tenth)
}
func BenchmarkTarjanSCCGnp_10_half(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_10_half)
}
func BenchmarkTarjanSCCGnp_100_half(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_100_half)
}
func BenchmarkTarjanSCCGnp_1000_half(b *testing.B) {
	benchmarkTarjanSCC(b, gnpDirected_1000_half)
}

func benchmarkDirectedCyclesIn(b *testing.B, g graph.Directed) {
	for i := 0; i < b.N; i++ {
		DirectedCyclesIn(g)
	}
}

func BenchmarkDirectedCyclesInGnp_10_tenth(b *testing.B) {
	benchmarkDirectedCyclesIn(b, gnpDirected_10_tenth)
}
func BenchmarkDirectedCyclesInGnp_10_half(b *testing.B) {
	benchmarkDirectedCyclesIn(b, gnpDirected_10_half)
}

func benchmarkUndirectedCyclesIn(b *testing.B, g graph.Undirected) {
	for i := 0; i < b.N; i++ {
		UndirectedCyclesIn(g)
	}
}

func BenchmarkUndirectedCyclesInGnp_10_tenth(b *testing.B) {
	benchmarkUndirectedCyclesIn(b, gnpUndirected_10_tenth)
}
func BenchmarkUndirectedCyclesInGnp_100_tenth(b *testing.B) {
	benchmarkUndirectedCyclesIn(b, gnpUndirected_100_tenth)
}
func BenchmarkUndirectedCyclesInGnp_10_half(b *testing.B) {
	benchmarkUndirectedCyclesIn(b, gnpUndirected_10_half)
}
func BenchmarkUndirectedCyclesInGnp_100_half(b *testing.B) {
	benchmarkUndirectedCyclesIn(b, gnpUndirected_100_half)
}
