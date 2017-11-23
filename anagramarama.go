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
	frequencyMap []byte

	// alternativeWords holds the list of alternatives for one candidate word.
	alternativeWords map[string][]string

	// workerRequest contains all the necessary information to start
	// a tree of anagrams (initial call to anawords).
	workerRequest struct {
		pmap     frequencyMap
		plen     int
		cand     []string
		base     []string
		altwords alternativeWords
		maxwords int
	}
)

// candidates reads a slice of words and produces a list of candidate and
// alternative words.  A candidate word is a word fully contained in the
// original phrase. Alternative words contain a slice of anagrams from the
// candidate word, keyed by the sorted characters of the candidate word.
func candidates(words []string, phrase string, minWordLen, maxWordLen int) ([]string, alternativeWords) {
	cand := []string{}
	altwords := alternativeWords{}

	pmap := make(frequencyMap, frequencyMapLen)
	freqmap(pmap, phrase)
	plen := len(phrase)

wordLoop:
	for _, w := range words {
		// Next word immediately if word is larger than phrase.
		if len(w) > plen {
			continue
		}
		// Reject word if outside desired word length limits
		if minWordLen != 0 && len(w) < minWordLen {
			continue
		}
		if maxWordLen != 0 && len(w) > maxWordLen {
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

// freqmap creates a frequency map of every letter in the word. It assumes only
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
	// We use a statically initialized overlay slice to avoid having to copy
	// the original frequencyMap. 0xff in a position means "not yet used".
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
// of candidate words. It will spawn a number of parallel goroutines to process
// each "root" (defined in parallelism.)
func anagrams(phrase string, cand []string, altwords alternativeWords, parallelism, maxwords int) []string {
	ret := []string{}

	// Pre-calculate frequency map and length of phrase, since it does not change.
	pmap := make(frequencyMap, frequencyMapLen)
	freqmap(pmap, phrase)
	plen := len(phrase)

	// Create request & response channels and start workers
	reqchan := make(chan workerRequest, parallelism)
	respchan := make(chan []string, parallelism)
	for i := 0; i < parallelism; i++ {
		go anaworker(reqchan, respchan)
	}

	pending := 0
	for ix, w := range cand {
		// Ignore removed candidate words.
		if w == "" {
			continue
		}
		// Immediately print candidates that match the len of phrase and remove them
		// from the slice, since they're anagrams by definition.
		if len(w) == plen {
			r := altCartesianProduct([]string{w}, altwords)
			ret = append(ret, r...)
			cand[ix] = ""
		}
		// Send the request to the pool of workers.
		req := workerRequest{
			pmap:     pmap,
			plen:     plen,
			cand:     cand[ix+1:],
			base:     []string{cand[ix]},
			altwords: altwords,
			maxwords: maxwords}
		reqchan <- req
		pending++

		r, nread := readResponses(respchan)
		if nread > 0 {
			ret = append(ret, r...)
			pending -= nread
		}
	}
	// Keep reading requests until no more pending requests exist.
	r := readNResponses(respchan, pending)
	ret = append(ret, r...)

	return ret
}

// anaworker continuously read a channel with the request of the work to be
// done and spawns anawords to recursively deal with it. The result from
// anawords is returned in the response channel.
func anaworker(req chan workerRequest, resp chan []string) {
	for {
		c := <-req
		ret := anawords(c.pmap, c.plen, c.cand, c.base, c.altwords, c.maxwords)
		resp <- ret
	}
}

// readResponses attempts reads all responses from the channel.  It returns the
// responses as a single slice of strings and the number of responses read. The
// function is non-blocking and will return when no response is available.
func readResponses(respchan chan []string) ([]string, int) {
	ret := []string{}
	nread := 0

	for {
		select {
		case r := <-respchan:
			nread++
			if len(r) > 0 {
				ret = append(ret, r...)
			}
		default:
			return ret, nread
		}
	}
}

// readNResponses attempts to read exactly N responses from the channel. If will
// block and wait on the channel if necessary.
func readNResponses(respchan chan []string, pending int) []string {
	ret := []string{}
	for ; pending > 0; pending-- {
		r := <-respchan
		if len(r) > 0 {
			ret = append(ret, r...)
		}
	}
	return ret
}

// anawords recursively generates a list of anagrams for the specified list of
// candidates, starting with 'base' as the root. This function may take an
// impossibly long time if the number of candidate words is too large.
func anawords(pmap frequencyMap, plen int, cand []string, base []string, altwords alternativeWords, maxwords int) []string {
	blen := 0
	for _, w := range base {
		blen += len(w)
	}
	//fmt.Printf("DEBUG: base=%q, blen=%d, plen=%d\n", base, blen, plen)

	// If length current base is longer than phrase, skip.
	if blen > plen {
		return []string{}
	}

	// Return if maximum number of words is exceeded (recursion depth.)
	if len(base) > maxwords {
		return []string{}
	}

	ret := []string{}

	// Our base phrase is still shorter than the phrase. We continue if our
	// base is still a candidate word of phrase.
	if blen < plen {
		if !mapContains(pmap, base...) {
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
				break
			}
			newbase := append(base, cword)

			r := anawords(pmap, plen, cand[ix+1:], newbase, altwords, maxwords)
			if len(r) > 0 {
				ret = append(ret, r...)
			}
		}
		return ret
	}

	// At this point, the length of the base string is the same as the phrase.
	// This means we *may* have a valid anagram and need to use mapEquals to
	// determine that.  We also use altCartesianProduct to generate all
	// possible transpositions of every candidate word by any available
	// alternatives.
	bmap := make(frequencyMap, frequencyMapLen)
	freqmap(bmap, base...)
	if !mapEquals(pmap, bmap) {
		return []string{}
	}
	ret = altCartesianProduct(base, altwords)
	return ret
}

// altCartesianProduct reads a slice of slice of strings, where each element in
// the inner slice are words in an expression. It then creates a new line with
// each of the words with present in 'altwords' replaced, returning the new,
// complete slice.
func altCartesianProduct(base []string, altwords alternativeWords) []string {
	// Lines is a slice of slices. First level is a line. Second level is a
	// slice of words.
	lines := make([][]string, 1, 10)
	lines[0] = base

	// Iterate over each word on base>
	for wordpos := range base {
		// Iterate over each line.
		numlines := len(lines)
		for ixline := 0; ixline < numlines; ixline++ {
			line := lines[ixline]
			word := line[wordpos]
			aws := altwords[sortString(word)]
			// Iterate over each alternate word and append variations to the
			// original slice at the 'idx' position. The cycle then repeats for
			// the new slice, starting at the next word.
			for _, aw := range aws {
				// Ignore the word itself.
				if aw == word {
					continue
				}
				newline := make([]string, len(line))
				for i := range line {
					newline[i] = line[i]
				}
				newline[wordpos] = aw
				lines = append(lines, newline)
			}
		}
	}

	// Convert [][]string into a []string with each word separated by spaces.
	ret := []string{}
	for _, line := range lines {
		ret = append(ret, strings.Join(line, " "))
	}
	return ret
}
