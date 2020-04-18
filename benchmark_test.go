// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package blobloom_test

import (
	"math/rand"
	"testing"
)

// These benchmarks simulate a situation where SHA-256 hashes are stored in a
// Bloom filter, using the first eight bytes as the Bloom filter hashes.

const hashSize = 32

func makehashes(n int, seed int64) []byte {
	h := make([]byte, n*hashSize)
	r := rand.New(rand.NewSource(seed))
	r.Read(h)

	return h
}

// In each iteration, add a SHA-256 into a Bloom filter with the given capacity
// and desired FPR.
func benchmarkAdd(b *testing.B, capacity int, fpr float64) {
	hashes := makehashes(b.N, 51251991517)
	f := newBF(capacity, fpr)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h := hashes[i*hashSize : (i+1)*hashSize]
		f.Add(h)
	}
}

func BenchmarkAdd1e5_1e2(b *testing.B) { benchmarkAdd(b, 1e5, 1e-2) }
func BenchmarkAdd1e6_1e2(b *testing.B) { benchmarkAdd(b, 1e6, 1e-2) }
func BenchmarkAdd1e7_1e2(b *testing.B) { benchmarkAdd(b, 1e7, 1e-2) }
func BenchmarkAdd1e8_1e2(b *testing.B) { benchmarkAdd(b, 1e8, 1e-2) }
func BenchmarkAdd1e5_1e3(b *testing.B) { benchmarkAdd(b, 1e5, 1e-3) }
func BenchmarkAdd1e6_1e3(b *testing.B) { benchmarkAdd(b, 1e6, 1e-3) }
func BenchmarkAdd1e7_1e3(b *testing.B) { benchmarkAdd(b, 1e7, 1e-3) }
func BenchmarkAdd1e8_1e3(b *testing.B) { benchmarkAdd(b, 1e8, 1e-3) }

// In each iteration, test for a SHA-256 in a Bloom filter with the given capacity
// and desired FPR that has that SHA-256 added to it.
func benchmarkTestPos(b *testing.B, capacity int, fpr float64) {
	hashes := makehashes(capacity, 0x5128351a)
	f := newBF(capacity, fpr)

	for i := 0; i < capacity; i++ {
		h := hashes[i*hashSize : (i+1)*hashSize]
		f.Add(h)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		j := i % capacity
		h := hashes[j*hashSize : (j+1)*hashSize]
		if !f.Has(h) {
			b.Fatalf("%x added to Bloom filter but not retrieved", h)
		}
	}
}

func BenchmarkTestPos1e5_1e2(b *testing.B) { benchmarkTestPos(b, 1e5, 1e-2) }
func BenchmarkTestPos1e6_1e2(b *testing.B) { benchmarkTestPos(b, 1e6, 1e-2) }
func BenchmarkTestPos1e7_1e2(b *testing.B) { benchmarkTestPos(b, 1e7, 1e-2) }
func BenchmarkTestPos1e8_1e2(b *testing.B) { benchmarkTestPos(b, 1e8, 1e-2) }
func BenchmarkTestPos1e5_1e3(b *testing.B) { benchmarkTestPos(b, 1e5, 1e-3) }
func BenchmarkTestPos1e6_1e3(b *testing.B) { benchmarkTestPos(b, 1e6, 1e-3) }
func BenchmarkTestPos1e7_1e3(b *testing.B) { benchmarkTestPos(b, 1e7, 1e-3) }
func BenchmarkTestPos1e8_1e3(b *testing.B) { benchmarkTestPos(b, 1e8, 1e-3) }

// In each iteration, test for the presence of a SHA-256 in a filled Bloom filter
// with the given capacity and desired FPR.
func benchmarkTestNeg(b *testing.B, capacity int, fpr float64) {
	r := rand.New(rand.NewSource(0xae694))
	f := newBF(capacity, fpr)

	h := make([]byte, hashSize)
	for i := 0; i < capacity; i++ {
		r.Read(h)
		f.Add(h)
	}

	// Make new hashes. Assume these are all distinct from the inserted ones.
	const ntest = 8192
	hashes := makehashes(ntest, 562175)

	b.ResetTimer()

	fp := 0
	for i := 0; i < b.N; i++ {
		j := i % ntest
		h := hashes[j*hashSize : (j+1)*hashSize]
		if f.Has(h) {
			fp++
		}
	}

	b.Logf("false positive rate = %.3f%%", 100*float64(fp)/float64(b.N))
}

func BenchmarkTestNeg1e5_1e2(b *testing.B) { benchmarkTestNeg(b, 1e5, 1e-2) }
func BenchmarkTestNeg1e6_1e2(b *testing.B) { benchmarkTestNeg(b, 1e6, 1e-2) }
func BenchmarkTestNeg1e7_1e2(b *testing.B) { benchmarkTestNeg(b, 1e7, 1e-2) }
func BenchmarkTestNeg1e8_1e2(b *testing.B) { benchmarkTestNeg(b, 1e8, 1e-2) }
func BenchmarkTestNeg1e5_1e3(b *testing.B) { benchmarkTestNeg(b, 1e5, 1e-3) }
func BenchmarkTestNeg1e6_1e3(b *testing.B) { benchmarkTestNeg(b, 1e6, 1e-3) }
func BenchmarkTestNeg1e7_1e3(b *testing.B) { benchmarkTestNeg(b, 1e7, 1e-3) }
func BenchmarkTestNeg1e8_1e3(b *testing.B) { benchmarkTestNeg(b, 1e8, 1e-3) }

// In each iteration, test for the presence of a SHA-256 in an empty Bloom filter
// with the given capacity and desired FPR.
func benchmarkTestEmpty(b *testing.B, capacity int, fpr float64) {
	hashes := makehashes(b.N, 054271)
	f := newBF(capacity, fpr)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		f.Has(hashes[i*hashSize : (i+1)*hashSize])
	}
}

func BenchmarkTestEmpty1e5_1e2(b *testing.B) { benchmarkTestEmpty(b, 1e5, 1e-2) }
func BenchmarkTestEmpty1e6_1e2(b *testing.B) { benchmarkTestEmpty(b, 1e6, 1e-2) }
func BenchmarkTestEmpty1e7_1e2(b *testing.B) { benchmarkTestEmpty(b, 1e7, 1e-2) }
func BenchmarkTestEmpty1e8_1e2(b *testing.B) { benchmarkTestEmpty(b, 1e8, 1e-2) }
func BenchmarkTestEmpty1e5_1e3(b *testing.B) { benchmarkTestEmpty(b, 1e5, 1e-3) }
func BenchmarkTestEmpty1e6_1e3(b *testing.B) { benchmarkTestEmpty(b, 1e6, 1e-3) }
func BenchmarkTestEmpty1e7_1e3(b *testing.B) { benchmarkTestEmpty(b, 1e7, 1e-3) }
func BenchmarkTestEmpty1e8_1e3(b *testing.B) { benchmarkTestEmpty(b, 1e8, 1e-3) }
