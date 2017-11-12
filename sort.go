// This file is part of anagramarama - A simple anagram generator in Go.
// (C) Oct/2017 by Marco Paganini <paganini@paganini.net>
//
// Check the main repository at http://github.com/marcopaganini/anagramarama
// for more details.
//
// Sort interfaces and functions.

package main

import (
	"sort"
)

func (x byLen) Len() int {
	return len(x)
}

func (x byLen) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x byLen) Less(i, j int) bool {
	leni := len(x[i])
	lenj := len(x[j])
	return leni < lenj
}

type sortRunes []rune

func (s sortRunes) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortRunes) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortRunes) Len() int {
	return len(s)
}

func sortString(s string) string {
	r := []rune(s)
	sort.Sort(sortRunes(r))
	return string(r)
}
