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
	frequencyMapLen = 26 // Uppercase letters.
)

type (
	// frequencyMap holds a letter to frequency map. Only 'frequencyMapLen'
	// characters are supported.
	frequencyMap [frequencyMapLen]int
)

// candidates reads a slice of words and produces a list of candidate words
// (i.e, words that could be anagrammed to our phrase).
func candidates(words []string, phrase string, minWordLen, maxWordLen int) []string {
	var cand []string
	pmap := freqmap(&phrase)
	plen := len(phrase)

wordLoop:
	for _, w := range words {
		wordlen := len(w)

		// Next word immediately if word is larger than phrase.
		if wordlen > plen {
			continue
		}
		// Reject word if outside desired word length limits
		if wordlen < minWordLen || wordlen > maxWordLen {
			continue
		}
		// Ignore anything not in [A-Z].
		w = strings.ToUpper(w)
		for _, r := range w {
			if r < 'A' || r > 'Z' {
				continue wordLoop
			}
		}
		if mapContains(&pmap, &w) {
			cand = append(cand, w)
		}
	}

	sort.Sort(byLen(cand))
	return cand
}

// freqmap creates a frequency map of every letter in the word. It assumes only
// uppercase letters as input and uses a slice instead of maps for performance
// reasons.
func freqmap(word *string) frequencyMap {
	var m frequencyMap

	for ix := 0; ix < len(*word); ix++ {
		r := (*word)[ix]
		idx := r - 'A'
		m[idx]++
	}
	return m
}

// mapLen returns the length of the map, in characters.
func mapLen(m frequencyMap) int {
	var size int
	for i := 0; i < frequencyMapLen; i++ {
		size += m[i]
	}
	return size
}

// mapContains returns true if map a contains the string.
func mapContains(a *frequencyMap, word *string) bool {
	var smap frequencyMap

	for i := 0; i < len(*word); i++ {
		idx := (*word)[i] - 'A'
		smap[idx]++
	}
	for i := 0; i < frequencyMapLen; i++ {
		if smap[i] > (*a)[i] {
			return false
		}
	}
	return true
}

// mapSubtract returns a map representing map a - map b.
func mapSubtract(m frequencyMap, words []string) frequencyMap {
	total := frequencyMap{}

	for i := 0; i < len(words); i++ {
		for j := 0; j < len(words[i]); j++ {
			idx := words[i][j] - 'A'
			total[idx]++
		}
	}
	for i := 0; i < frequencyMapLen; i++ {
		total[i] = m[i] - total[i]
	}
	return total
}

// mapIsEmpty returns true if the map is empty, false otherwise.
func mapIsEmpty(m frequencyMap) bool {
	for i := 0; i < frequencyMapLen; i++ {
		if m[i] > 0 {
			return false
		}
	}
	return true
}

// anagrams recursively generates a list of anagrams for the specified list of
// candidates, starting with 'base' as the root. If 'depth' is specified,
// recursion will stop at this level. This essentially limits the number of
// words in an anagram. This function may take an impossibly long time if the
// number of candidate words is too large.
func anagrams(pmap frequencyMap, cand []string, base []string, numwords, maxwords int) []string {
	var ret []string

	// maximum recursion depth (number of words)
	if numwords > maxwords {
		return nil
	}
	numwords++

	//fmt.Printf("Got base=%s\n", base)
	leftmap := mapSubtract(pmap, base)

	// Perfect match.
	if mapIsEmpty(leftmap) {
		return append(ret, strings.Join(base, " "))
	}

	charsleft := mapLen(leftmap)

	for ix := 0; ix < len(cand); ix++ {
		cword := cand[ix]
		// The input list of words is sorted by word length.  If we the length
		// of the current base + the current word exceeds the total length of
		// the phrase, no more anagrams exist from this point on.
		if len(cword) > charsleft {
			break
		}

		// Only recurse if cword fits the remaining characters.
		if !mapContains(&leftmap, &cword) {
			continue
		}

		// New base is our current base + new word.
		newbase := append(base, cword)
		r := anagrams(pmap, cand[ix+1:], newbase, numwords, maxwords)
		if r != nil {
			ret = append(ret, r...)
		}
	}
	return ret
}
