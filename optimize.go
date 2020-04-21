// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blobloom

import "math"

// A Config holds parameters for Optimize or NewOptimized.
type Config struct {
	// Capacity is the expected number of distinct keys to be added.
	// More keys can always be added, but the false positive rate can be
	// expected to drop below FPRate is the number exceeds the Capacity.
	Capacity uint64

	// Desired lower bound on the false positive rate when the Bloom filter
	// has been filled to capacity.
	FPRate float64

	// Maximum size of the Bloom filter in bits. Zero means no limit.
	MaxBits uint64

	// Trigger the "contains filtered or unexported fields" message for
	// forward compatibility and to force the caller to use named fields.
	_ struct{}
}

// NewOptimized is shorthand for New(Optimize(cfg)).
func NewOptimized(cfg Config) *Filter {
	return New(Optimize(cfg))
}

// Optimize returns numbers of keys and hash functions that achieve the
// desired false positive described by cfg.
//
// The estimated number of bits is imprecise for false positives rates below
// ca. 1e-15.
func Optimize(cfg Config) (nbits uint64, nhashes int) {
	var (
		n = float64(cfg.Capacity)
		p = cfg.FPRate
	)

	if p <= 0 || p > 1 {
		panic("false positive rate for a Bloom filter must be > 0, <= 1")
	}
	if n == 0 {
		// Assume the client wants to add at least one key; log2(0) = -inf.
		n = 1
	}

	// The optimal nbits/n is c = -log2(p) / ln(2) for a vanilla Bloom filter.
	c := math.Ceil(-math.Log2(p) / math.Ln2)
	if c < float64(len(correctC)) {
		c = float64(correctC[int(c)])
	} else {
		// We can't achieve the desired FPR. Just triple the number of bits.
		c *= 3
	}
	nbits = uint64(c * n)

	// Round up to a multiple of BlockBits.
	if nbits%BlockBits != 0 {
		nbits += BlockBits - nbits%BlockBits
	}

	maxbits := uint64(1<<32) * BlockBits
	if cfg.MaxBits != 0 && cfg.MaxBits < maxbits {
		maxbits = cfg.MaxBits
	}
	if nbits > maxbits {
		nbits = maxbits
		// Round down to a multiple of BlockBits.
		nbits -= nbits % BlockBits
	}

	// The corresponding optimal number of hash functions is k = c * log(2).
	// Try rounding up and down to see which rounding is better.
	// XXX Should we use the full FPR here instead of block FPR?
	c = float64(nbits) / n
	k := c * math.Ln2
	if logFprBlock(c, math.Floor(k)) > logFprBlock(c, math.Ceil(k)) {
		k = math.Floor(k)
	} else {
		k = math.Ceil(k)
	}
	nhashes = int(k)

	if nhashes < 1 {
		nhashes = 1
	}

	return nbits, nhashes
}

// correctC maps c = m/n for a vanilla Bloom filter to the c' for a
// blocked Bloom filter.
//
// This is Putze et al.'s Table I, extended down to zero.
// For c > 34, the values become huge and are hard to compute.
var correctC = []byte{
	1, 1, 2, 4, 5,
	6, 7, 8, 9, 10, 11, 12, 13, 14, 16, 17, 18, 20, 21, 23,
	25, 26, 28, 30, 32, 35, 38, 40, 44, 48, 51, 58, 64, 74, 90,
}

// FPRate computes an estimate of the false positive rate of a Bloom filter
// after nkeys distinct keys have been added.
func FPRate(nkeys, nbits uint64, nhashes int) float64 {
	if nkeys == 0 {
		return 0
	}
	return fpRate(float64(nbits)/float64(nkeys), float64(nhashes))
}

func fpRate(c, k float64) float64 {
	// Putze et al.'s Equation (3).
	var sum float64
	for i := float64(1); ; i++ {
		add := math.Exp(logPoisson(BlockBits/c, i) + logFprBlock(BlockBits/i, k))
		sum += add
		if add/sum < 1e-8 {
			break
		}
	}

	return sum
}

// FPRate computes an estimate of f's false positive rate after nkeys distinct
// keys have been added.
func (f *Filter) FPRate(nkeys uint64) float64 {
	return FPRate(nkeys, f.NumBits(), f.k)
}

// Log of the FPR of a single block.
func logFprBlock(c, k float64) float64 {
	return k * math.Log1p(-math.Exp(-k/c))
}

// Log of the Poisson distribution's pmf.
func logPoisson(λ, k float64) float64 {
	lg, _ := math.Lgamma(k + 1)
	return k*math.Log(λ) - λ - lg
}
