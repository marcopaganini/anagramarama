// This file is part of anagramarama - A simple anagram generator in Go.
// (C) Oct/2017 by Marco Paganini <paganini@paganini.net>
//
// Check the main repository at http://github.com/marcopaganini/anagramarama
// for more details.
//
// Sort interfaces and functions.

package main

// byLen defines a type to sort a slice of strings by
// the length of each element.
type byLen []string

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
