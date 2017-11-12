// This file is part of anagramarama - A simple anagram generator in Go.
//
// (C) Oct/2017 by Marco Paganini <paganini@paganini.net>
//
// Check the main repository at http://github.com/marcopaganini/anagramarama
// for more details.

package main

import (
	//"fmt"
	"sort"
	"strings"
)

const (
	frequencyMapLen = 26    // Uppercase letters.
	anagramCapacity = 30000 // Initial capacity of the slice to hold anagrams.
)

type (
	frequencyMap     []byte
	alternativeWords map[string][]string
	byLen            []string
)

// candidates reads a slice of words and produces a list of candidate and
// alternative words.  A candidate word is a word fully contained in the
// original phrase. Alternative words contain a slice of anagrams from the
// candidate word, keyed by the sorted characters of the word.
func candidates(words []string, phrase string) ([]string, alternativeWords) {
	cand := []string{}
	altwords := alternativeWords{}

	pmap := make(frequencyMap, frequencyMapLen)
	freqmap(pmap, phrase)
	plen := nonSpaceLen(phrase)

wordLoop:
	for _, w := range words {
		// Next word immediately if word is larger than phrase.
		if len(w) > plen {
			continue
		}
		// Ignore anything not in [A-Z].
		w := strings.ToUpper(w)
		for _, r := range w {
			if r < 'A' || r > 'Z' {
				continue wordLoop
			}
		}
		if mapContains(pmap, w) {
			sortword := sortString(w)
			if _, ok := altwords[sortword]; ok {
				altwords[sortword] = append(altwords[sortword], w)
			} else {
				altwords[sortword] = []string{w}
				cand = append(cand, w)
			}
		}
	}

	sort.Sort(byLen(cand))
	return cand, altwords
}

// freqmap creates a frequency map of every letter in the word. Assumes only
// uppercase letters as input and uses a slice instead of maps for performance
// reasons.
func freqmap(fm frequencyMap, str ...string) {
	for _, s := range str {
		for _, r := range s {
			idx := int(r) - int('A')
			fm[idx]++
		}
	}
}

// mapContains returns true if a string is fully contained in a frequency map.
func mapContains(a frequencyMap, str ...string) bool {
	overlay := frequencyMap{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ret := true
	for _, s := range str {
		for _, r := range s {
			idx := int(r) - int('A')
			if overlay[idx] == 0xff {
				overlay[idx] = a[idx]
			}
			if overlay[idx] == 0 {
				ret = false
				break
			}
			overlay[idx]--
		}
	}
	return ret
}

// mapEquals returns true if two frequency maps are identical.
func mapEquals(a, b frequencyMap) bool {
	for ix := 0; ix < frequencyMapLen; ix++ {
		if a[ix] != b[ix] {
			return false
		}
	}
	return true
}

// anagrams starts the recursive anagramming function for each word in the list
// of candidate words.
func anagrams(phrase string, cand []string, altwords alternativeWords) []string {
	ret := make([]string, 0, anagramCapacity)

	// Pre-calculate frequency map and length of phrase, since it does not change.
	pmap := make(frequencyMap, frequencyMapLen)
	freqmap(pmap, phrase)
	plen := nonSpaceLen(phrase)

	// Immediately print candidates that match the len of phrase and remove them
	// from the slice, since they're anagrams by definition.
	for ix, w := range cand {
		if len(w) == plen {
			r := multiReplace([]string{w}, altwords)
			ret = append(ret, r...)
			cand[ix] = ""
		}
	}

	for ix, w := range cand {
		if w == "" {
			continue
		}
		//fmt.Printf("Anagrams trying with base=%q\n", cand[ix])
		r := anawords(pmap, plen, cand[ix+1:], []string{cand[ix]}, altwords)
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
func anawords(pmap frequencyMap, plen int, cand []string, base []string, altwords alternativeWords) []string {
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

			r := anawords(pmap, plen, cand[ix+1:], newbase, altwords)

			if len(r) > 0 {
				//fmt.Printf("DEBUG: Appending %q to the list of anagrams\n", r)
				ret = append(ret, r...)
			}
		}
		//fmt.Printf("DEBUG: Returning %q now\n", ret)
		return ret
	}

	// The length of the base string is the same as the phrase. If we have
	// an exact match then we have an anagram.
	bmap := make(frequencyMap, frequencyMapLen)
	freqmap(bmap, base...)
	if !mapEquals(pmap, bmap) {
		//fmt.Println("DEBUG: no exact match for", base)
		return []string{}
	}
	ret = multiReplace(base, altwords)
	//fmt.Printf("Returning exact match: %q\n", ret)
	return ret
}

func multiReplace(base []string, altwords alternativeWords) []string {
	// Lines is a slice of slices. First level is a line. Second
	// level is a slice of words.
	lines := make([][]string, 1, 10)
	lines[0] = base

	// Iterate over each word on base>
	for wordpos := range base {
		//fmt.Printf("wordpos = %d\n", wordpos)
		// Iterate over each line.
		numlines := len(lines)
		for ixline := 0; ixline < numlines; ixline++ {
			line := lines[ixline]
			//fmt.Printf("Line is %q\n", line)
			word := line[wordpos]
			//fmt.Printf("Word is %q\n", word)
			aws := altwords[sortString(word)]
			//fmt.Printf("Alternative words: %q\n", aws)
			// Iterate over each alternate word and append variations to
			// the original slice at the 'idx' position. The cycle then
			// repeats for the new slice, starting at the next word.
			for _, aw := range aws {
				//fmt.Printf("Alternate word is: %q\n", aw)
				if aw == word {
					//fmt.Printf("Ignoring...\n")
					continue
				}
				newline := make([]string, len(line))
				for i := range line {
					newline[i] = line[i]
				}
				newline[wordpos] = aw
				//fmt.Printf("Adding %q to lines\n", newline)
				lines = append(lines, newline)
			}
		}
	}

	// Convert [][]string into a []string with each word separated by spaces.
	ret := []string{}
	for _, line := range lines {
		ret = append(ret, strings.Join(line, " "))
	}
	//fmt.Printf("Replaced to %q\n", ret)
	return ret
}
