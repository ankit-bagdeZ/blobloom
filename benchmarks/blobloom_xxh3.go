// Copyright 2020 the Blobloom authors
//
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

// +build xxh3

package benchmarks

import (
	"github.com/greatroar/blobloom"
	"github.com/zeebo/xxh3"
)

type bloomFilter blobloom.Filter

func (f *bloomFilter) Add(hash []byte) {
	h := xxh3.Hash(hash)
	((*blobloom.Filter)(f)).Add(h)
}

func (f *bloomFilter) Has(hash []byte) bool {
	h := xxh3.Hash(hash)
	return ((*blobloom.Filter)(f)).Has(h)
}

func newBF(capacity int, fpr float64) *bloomFilter {
	f := blobloom.NewOptimized(blobloom.Config{
		Capacity: uint64(capacity),
		FPRate:   fpr,
	})
	return (*bloomFilter)(f)
}
