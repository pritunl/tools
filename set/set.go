// Copyright (c) 2014 Dropbox, Inc.
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// 1. Redistributions of source code must retain the above copyright notice, this
// list of conditions and the following disclaimer.

// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.

// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package set implements an unordered collection of unique elements with
// the usual membership, iteration and binary set operations.
//
// Two implementations are provided. NewSet compares elements using Go
// equality, so elements must be comparable. NewKeyedSet derives a
// comparable key from each element with a caller-supplied function, which
// allows non-comparable or structurally-equal elements to be stored.
//
//	s := set.NewSet("a", "b")
//	s.Add("c")
//	if s.Contains("b") { ... }
//
// Sets are not safe for concurrent use.
//
// This package is derived from the Dropbox godropbox set package.
package set

// Set is an unordered collection of unique elements which supports lookups,
// insertions, deletions, iteration and common binary set operations. It is
// not safe for concurrent use.
//
// Methods that take another Set as an argument accept a nil Set, which is
// treated as empty.
type Set interface {
	// New returns a new empty Set of the same type and, for keyed sets,
	// with the same key function.
	New() Set

	// Copy returns a new Set containing exactly the same elements as this
	// set.
	Copy() Set

	// Len returns the number of elements in this set.
	Len() int

	// Contains reports whether v is an element of this set.
	Contains(v interface{}) bool
	// Add inserts v into this set. Adding an element that is already
	// present has no effect.
	Add(v interface{})
	// Remove deletes v from this set if it is present and reports whether
	// it was present.
	Remove(v interface{}) bool

	// Do calls f for every element of this set in unspecified order. The
	// behavior is undefined if f mutates this set.
	Do(f func(interface{}))
	// DoWhile calls f for every element of this set in unspecified order,
	// stopping early if f returns false. The behavior is undefined if f
	// mutates this set.
	DoWhile(f func(interface{}) bool)
	// Iter returns a channel from which each element of this set can be
	// received exactly once. The channel is closed after the last element.
	// The channel is fed by a goroutine, so callers must drain it to avoid
	// a leak. The values received are undefined if this set is mutated
	// before the channel is drained.
	Iter() <-chan interface{}

	// Union adds every element of s to this set.
	Union(s Set)
	// Intersect removes every element from this set that is not in s.
	Intersect(s Set)
	// Subtract removes every element of s from this set.
	Subtract(s Set)
	// Init removes all elements from this set.
	Init()
	// IsSubset reports whether every element of this set is in s.
	IsSubset(s Set) bool
	// IsSuperset reports whether every element of s is in this set.
	IsSuperset(s Set) bool
	// IsEqual reports whether this set and s contain exactly the same
	// elements.
	IsEqual(s Set) bool
	// RemoveIf removes every element v of this set for which f(v) is true.
	RemoveIf(f func(interface{}) bool)
}

// Union returns a new set containing every element of s1 and s2. Neither
// argument is modified. If s1 is nil a copy of s2 is returned, and if both
// are nil the result is nil.
func Union(s1 Set, s2 Set) Set {
	if s1 == nil {
		if s2 == nil {
			return nil
		}

		return s2.Copy()
	}
	s3 := s1.Copy()
	s3.Union(s2)
	return s3
}

// Intersect returns a new set containing the elements common to s1 and s2.
// Neither argument is modified. If s1 is nil an empty set of s2's type is
// returned, and if both are nil the result is nil.
func Intersect(s1 Set, s2 Set) Set {
	if s1 == nil {
		if s2 == nil {
			return nil
		}

		return s2.New()
	}
	s3 := s1.Copy()
	s3.Intersect(s2)
	return s3
}

// Subtract returns a new set containing the elements of s1 that are not in
// s2. Neither argument is modified. If s1 is nil an empty set of s2's type
// is returned, and if both are nil the result is nil.
func Subtract(s1 Set, s2 Set) Set {
	if s1 == nil {
		if s2 == nil {
			return nil
		}

		return s2.New()
	}
	s3 := s1.Copy()
	s3.Subtract(s2)
	return s3
}

// NewSet returns a new Set containing items. Elements are compared using
// Go equality, so every element must be a comparable type; adding a
// non-comparable value such as a slice or map panics.
func NewSet(items ...interface{}) Set {
	res := setImpl{
		data: make(map[interface{}]struct{}),
	}
	for _, item := range items {
		res.Add(item)
	}
	return res
}

// NewKeyedSet returns a new Set containing items in which element identity
// is determined by keyf. Two elements are considered the same when keyf
// returns equal keys for them, and the key must be a comparable type. When
// an element is added whose key is already present, the stored element is
// replaced by the new one. Iteration yields the stored elements, not their
// keys.
func NewKeyedSet(keyf func(interface{}) interface{}, items ...interface{}) Set {
	res := keyedSetImpl{
		data:    make(map[interface{}]interface{}),
		keyfunc: keyf,
	}
	for _, item := range items {
		res.Add(item)
	}
	return res
}

// setImpl is the Set implementation returned by NewSet. It stores elements
// directly as map keys.
type setImpl struct {
	data map[interface{}]struct{}
}

func (s setImpl) Len() int {
	return len(s.data)
}

func (s setImpl) New() Set {
	return NewSet()
}

func (s setImpl) Copy() Set {
	return copySet(s)
}

func (s setImpl) Init() {
	s.data = make(map[interface{}]struct{})
}

func (s setImpl) Contains(v interface{}) bool {
	_, ok := s.data[v]
	return ok
}

func (s setImpl) Add(v interface{}) {
	s.data[v] = struct{}{}
}

func (s setImpl) Remove(v interface{}) bool {
	_, ok := s.data[v]
	if ok {
		delete(s.data, v)
	}
	return ok
}

func (s setImpl) Do(f func(interface{})) {
	for key := range s.data {
		f(key)
	}
}

func (s setImpl) DoWhile(f func(interface{}) bool) {
	for key := range s.data {
		if !f(key) {
			break
		}
	}
}

func (s setImpl) Iter() <-chan interface{} {
	iter := make(chan interface{})
	go func() {
		for key := range s.data {
			iter <- key
		}
		close(iter)
	}()
	return iter
}

func (s setImpl) Union(s2 Set) {
	union(s, s2)
}

func (s setImpl) Intersect(s2 Set) {
	var toRemove []interface{}
	for key := range s.data {
		if s2 == nil || !s2.Contains(key) {
			toRemove = append(toRemove, key)
		}
	}

	for _, key := range toRemove {
		s.Remove(key)
	}
}

func (s setImpl) Subtract(s2 Set) {
	subtract(s, s2)
}

func (s setImpl) IsSubset(s2 Set) (isSubset bool) {
	return subset(s, s2)
}

func (s setImpl) IsSuperset(s2 Set) bool {
	return superset(s, s2)
}

func (s setImpl) IsEqual(s2 Set) bool {
	return equal(s, s2)
}

func (s setImpl) RemoveIf(f func(interface{}) bool) {
	var toRemove []interface{}
	for item := range s.data {
		if f(item) {
			toRemove = append(toRemove, item)
		}
	}

	for _, item := range toRemove {
		s.Remove(item)
	}
}

// keyedSetImpl is the Set implementation returned by NewKeyedSet. It stores
// elements as map values indexed by the key derived from keyfunc.
type keyedSetImpl struct {
	data    map[interface{}]interface{}
	keyfunc (func(interface{}) interface{})
}

func (s keyedSetImpl) Len() int {
	return len(s.data)
}

func (s keyedSetImpl) New() Set {
	return NewKeyedSet(s.keyfunc)
}

func (s keyedSetImpl) Copy() Set {
	return copySet(s)
}

func (s keyedSetImpl) Init() {
	s.data = make(map[interface{}]interface{})
}

func (s keyedSetImpl) Contains(v interface{}) bool {
	_, ok := s.data[s.keyfunc(v)]
	return ok
}

func (s keyedSetImpl) Add(v interface{}) {
	s.data[s.keyfunc(v)] = v
}

func (s keyedSetImpl) Remove(v interface{}) bool {
	key := s.keyfunc(v)
	_, ok := s.data[key]
	if ok {
		delete(s.data, key)
	}
	return ok

}

func (s keyedSetImpl) Do(f func(interface{})) {
	for _, v := range s.data {
		f(v)
	}
}

func (s keyedSetImpl) DoWhile(f func(interface{}) bool) {
	for _, v := range s.data {
		if !f(v) {
			break
		}
	}
}

func (s keyedSetImpl) Iter() <-chan interface{} {
	iter := make(chan interface{})
	go func() {
		for _, v := range s.data {
			iter <- v
		}
		close(iter)
	}()
	return iter
}

func (s keyedSetImpl) Union(s2 Set) {
	union(s, s2)
}

func (s keyedSetImpl) Intersect(s2 Set) {
	var toRemove []interface{}
	for _, v := range s.data {
		if s2 == nil || !s2.Contains(v) {
			toRemove = append(toRemove, v)
		}
	}

	for _, v := range toRemove {
		s.Remove(v)
	}
}

func (s keyedSetImpl) Subtract(s2 Set) {
	subtract(s, s2)
}

func (s keyedSetImpl) IsSubset(s2 Set) (isSubset bool) {
	return subset(s, s2)
}

func (s keyedSetImpl) IsSuperset(s2 Set) bool {
	return superset(s, s2)
}

func (s keyedSetImpl) IsEqual(s2 Set) bool {
	return equal(s, s2)
}

func (s keyedSetImpl) RemoveIf(f func(interface{}) bool) {
	var toRemove []interface{}
	for _, item := range s.data {
		if f(item) {
			toRemove = append(toRemove, item)
		}
	}

	for _, item := range toRemove {
		s.Remove(item)
	}
}

// The functions below are shared by both implementations and operate only
// through the Set interface.

// equal reports whether s and s2 contain exactly the same elements.
func equal(s Set, s2 Set) bool {
	if s.Len() != s2.Len() {
		return false
	}

	return s.IsSubset(s2)
}

// superset reports whether every element of s2 is in s.
func superset(s Set, s2 Set) bool {
	return s2.IsSubset(s)
}

// subset reports whether every element of s is in s2.
func subset(s Set, s2 Set) (isSubset bool) {
	isSubset = true
	s.DoWhile(func(item interface{}) bool {
		if !s2.Contains(item) {
			isSubset = false
		}
		return isSubset
	})
	return
}

// subtract removes every element of s2 from s. A nil s2 is a no-op.
func subtract(s Set, s2 Set) {
	if s2 == nil {
		return
	}
	s2.Do(func(item interface{}) { s.Remove(item) })
}

// union adds every element of s2 to s. A nil s2 is a no-op.
func union(s Set, s2 Set) {
	if s2 == nil {
		return
	}
	s2.Do(func(item interface{}) { s.Add(item) })
}

// copySet returns a new set of the same type as s containing its elements.
func copySet(s Set) Set {
	res := s.New()
	res.Union(s)
	return res
}
