// Copyright (c) 2016, German Neuroinformatics Node (G-Node),
//                     Adrian Stoewer <adrian.stoewer@rz.ifi.lmu.de>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted under the terms of the BSD License. See
// LICENSE file in the root of the Project.

package util

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var arrayRegex = regexp.MustCompile(`((((([^",\\{}\s]|NULL)+|"([^"\\]|\\"|\\\\)*")))(,)?)`)
var regexValIndex = 3

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// ToSnakeCase turns a CamelCase string into its snake_case equivalent.
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// NewStringSet creates a new StringSet from a slice of strings.
func NewStringSet(strs ...string) StringSet {
	set := StringSet{}
	for _, s := range strs {
		set[s] = true
	}
	return set
}

// StringSet is a simple set implementation based on map.
type StringSet map[string]bool

// Add returns a set with one additional element.
func (set StringSet) Add(s string) StringSet {
	if !set.Contains(s) {
		fresh := StringSet{s: true}
		for s := range set {
			fresh[s] = true
		}
		return fresh
	}
	return set
}

// Contains returns true if a string is part of the set.
func (set StringSet) Contains(s string) bool {
	contains := set[s]
	return contains
}

// IsSuperset returns true if the set is a superset of the other set.
func (set StringSet) IsSuperset(other StringSet) bool {
	if set.Len() == 0 {
		return false
	}
	for _, s := range other.Strings() {
		if !set.Contains(s) {
			return false
		}
	}
	return true
}

// Union returns a set that contains all elements from both sets.
func (set StringSet) Union(other StringSet) StringSet {
	switch other.Len() {
	case 0:
		return set
	case 1:
		return set.Add(other.Strings()[0])
	default:
		fresh := set
		isCopy := false
		for _, s1 := range other.Strings() {
			if !set.Contains(s1) {
				if !isCopy {
					fresh := StringSet{}
					for s2 := range set {
						fresh[s2] = true
					}
				}
				fresh[s1] = true
			}
		}
		return fresh
	}
}

// Intersect returns a new set containing only the elements contained in both
// original sets
func (set StringSet) Intersect(other StringSet) StringSet {
	a, b := set, other
	if a.Len() > b.Len() {
		a, b = b, a
	}
	inter := make([]string, 0, a.Len())
	for s := range a {
		if b.Contains(s) {
			inter = append(inter, s)
		}
	}
	return NewStringSet(inter...)
}

// Len returns the number of elements in the set.
func (set StringSet) Len() int {
	return len(set)
}

// Strings returns all elements as slice of strings.
func (set StringSet) Strings() []string {
	strs := make([]string, 0, len(set))
	for s := range set {
		strs = append(strs, s)
	}
	sort.StringSlice(strs).Sort()
	return strs
}

// Scan implements the Scanner interface.
func (set *StringSet) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return errors.New("Souce wan not []byte")
	}

	asStr := string(asBytes)
	results := make([]string, 0)
	matches := arrayRegex.FindAllStringSubmatch(asStr, -1)
	for _, match := range matches {
		s := match[regexValIndex]
		s = strings.Trim(s, "\"")
		results = append(results, s)
	}

	(*set) = NewStringSet(results...)
	return nil
}

// Value implements the driver Valuer interface.
func (set StringSet) Value() (driver.Value, error) {
	if set.Len() == 0 {
		return `{}`, nil
	}

	quoted := make([]string, set.Len(), set.Len())
	for i, s := range set.Strings() {
		quoted[i] = fmt.Sprintf(strconv.Quote(s))
	}

	return fmt.Sprintf("{%s}", strings.Join(quoted, ",")), nil
}

// Difference returns a new util.StringSet containing all elements
// which are not found in the other util.StringSet.
func (set StringSet) Difference(other StringSet) StringSet {
	ret := make([]string, 0, set.Len())
	for k := range set {
		if !other.Contains(k) {
			ret = append(ret, k)
		}
	}
	return NewStringSet(ret...)
}
