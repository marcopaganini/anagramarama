// This file is part of anagramarama - A simple anagram generator in Go.
//
// (C) Oct/2017 by Marco Paganini <paganini@paganini.net>
//
// Check the main repository at http://github.com/marcopaganini/anagramarama
// for more details.

package main

import (
	"sort"
	"strings"
)

const (
	frequencyMapLen = 26 // Uppercase letters
)

type frequencyMap []int

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

// candidates reads a slice of words and produces a list of candidate words.
// All words are converted to uppercase when read. Words containing non-letter
// characters are silently ignored.
func candidates(words []string, phrase string) []string {
	cand := []string{}

	pmap := make(frequencyMap, frequencyMapLen)
	freqmap(pmap, phrase)
	plen := nonSpaceLen(phrase)

wordLoop:
	for _, w := range words {
		// Next word immediately if word is larger than phrase.
		if len(w) > plen {
			continue
		}
		w := strings.ToUpper(w)
		for _, r := range w {
			if r < 'A' || r > 'Z' {
				continue wordLoop
			}
		}
		if mapContains(pmap, w) {
			cand = append(cand, w)
		}
	}

	sort.Sort(byLen(cand))
	return cand
}

// freqmap creates a frequency map of every letter in the word. Assumes only
// uppercase letters as input and uses a slice instead of maps for performance
// reasons.
func freqmap(fm frequencyMap, str ...string) {
	for _, s := range str {
		for _, r := range s {
			idx := int(r) - int('A')
			if idx < 0 || idx > 25 {
				continue
			}
			fm[int(r)-int('A')]++
		}
	}
}

// mapContains returns true if a string is fully contained in a frequency map.
func mapContains(a frequencyMap, str ...string) bool {
	acopy := make(frequencyMap, frequencyMapLen)
	copy(acopy, a)

	//fmt.Printf("==== DEBUG ====\nmapcontains string slice: %q\n", str)

	for _, s := range str {
		//fmt.Printf("DEBUG: mapcontains string %q\n%q\n", s, acopy)
		for _, r := range s {
			idx := int(r) - int('A')
			if idx < 0 || idx > 25 {
				continue
			}
			if acopy[idx] == 0 {
				//fmt.Printf("DEBUG: idx=%d, count=%d, returning false\n", idx, acopy[idx])
				return false
			}
			acopy[idx]--
		}
	}
	//fmt.Printf("DEBUG: mapcontains returning true\n")
	return true
}

// mapEquals returns true if two frequency maps are identical.
func mapEquals(a, b frequencyMap) bool {
	if len(a) != len(b) {
		return false
	}
	for ix := 0; ix < frequencyMapLen; ix++ {
		if a[ix] != b[ix] {
			return false
		}
	}
	return true
}

// anagrams starts the recursive anagramming function for each word in the list
// of candidate words.
func anagrams(phrase string, cand []string) []string {
	ret := []string{}
	pmap := make(frequencyMap, frequencyMapLen)
	freqmap(pmap, phrase)
	plen := nonSpaceLen(phrase)

	// Immediately print candidates that match the len of phrase and remove them
	// from the slice, since they're anagrams by definition.
	for ix, w := range cand {
		if len(w) == plen {
			ret = append(ret, w)
			cand[ix] = ""
		}
	}

	for ix, w := range cand {
		if w == "" {
			continue
		}
		//fmt.Printf("Anagrams trying with base=%q\n", cand[ix])
		r := anawords(pmap, plen, cand[ix+1:], []string{cand[ix]})
		if len(r) > 0 {
			ret = append(ret, r...)
		}
	}
	return ret
}

// noSpaceLen returns the length of the string ignoring spaces.
func nonSpaceLen(s string) int {
	c := 0
	for _, r := range s {
		if r != ' ' {
			c++
		}
	}
	return c
}

// anawords recursively generates a list of anagrams for the specified list of
// candidates, starting with 'base' as the root.
func anawords(pmap frequencyMap, plen int, cand []string, base []string) []string {
	blen := 0
	for _, w := range base {
		blen += len(w)
	}
	//fmt.Printf("DEBUG: base=%q, blen=%d, plen=%d\n", base, blen, plen)
	//fmt.Printf("DEBUG: cand=%q\n", cand)

	// If current base is longer than phrase, skip.
	if blen > plen {
		//fmt.Println("DEBUG: Base too long:", base)
		return []string{}
	}

	ret := []string{}

	// Our base phrase is still shorter than the phrase. We continue if our
	// base is still a candidate word of phrase.
	if blen < plen {
		if !mapContains(pmap, base...) {
			//fmt.Printf("DEBUG: Map does not contain base: %q\n", base)
			return []string{}
		}
		// Recurse with each word on the list of candidates and our base.
		for ix, cword := range cand {
			// Ignore removed (blank) words.
			if cword == "" {
				continue
			}
			// Optimization: the input list of words is sorted by word length.
			// If we the length of the current base + the current word exceeds
			// the total length of the phrase, we can return immediately, since
			// all further executions will be invalid.
			if blen+len(cword) > plen {
				//fmt.Printf("DEBUG: skipping since base is too large: blen=%d, cwordlen=%d, cword=%q\n", blen, len(cword), cword)
				break
			}
			newbase := append(base, cword)
			r := anawords(pmap, plen, cand[ix+1:], newbase)
			if len(r) > 0 {
				//fmt.Printf("DEBUG: Appending %q to the list of anagrams\n", r)
				ret = append(ret, r...)
			}
		}
		//fmt.Printf("DEBUG: Returning %q now\n", ret)
		return ret
	}

	// Need an exact match here.
	//fmt.Println("DEBUG: Need an exact match for", base)
	bmap := make(frequencyMap, frequencyMapLen)
	freqmap(bmap, base...)
	if !mapEquals(pmap, bmap) {
		//fmt.Println("DEBUG: no exact match for", base)
		return []string{}
	}
	//fmt.Printf("DEBUG: Yay got match %q\n", base)
	return []string{strings.Join(base, " ")}
}
