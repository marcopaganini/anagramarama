// anagramarama - A simple anagram generator in Go.
//
// This is a simple (and inefficient) anagram generator in Go. Its main purpose
// is not to provide a platform for people learning recursive algorithms and
// efficiency tweaks in Go.
//
// A few possible improvements:
//
// 1) How to make it more efficient in general? Where is time being consumed?
// 2) How can we make subtract() faster?
// 3) anagram() can be optimized.
// 4) Make the program UTF-8 safe.
// 5) Speed up loading by saving a ready made version of wordLetters.
//
// Also, it's quite possible many bug exist.
//
// (C) Oct/2017 by Marco Paganini <paganini@paganini.net>
//
// Check the main repository at http://github.com/marcopaganini/anagramarama
// for more details.

package main

import (
	"bufio"
	"io"
	"sort"
	"strings"
)

// wordLetters maps a string with the sorted letters of a word to
// the list of words containing exactly those letters.
type wordLetters map[string][]string

// sortedString returns the input string with the letters sorted.
func sortedString(s string, filter map[rune]bool) string {
	letters := []string{}
	for _, r := range s {
		if _, ok := filter[r]; !ok {
			letters = append(letters, string(r))
		}
	}
	sort.Strings(letters)
	return strings.Join(letters, "")
}

// addUniqueWord adds a new word to the slice of sortedLetters,
// avoiding duplicates.
func addUniqueWord(words wordLetters, sortedLetters, word string) {
	_, ok := words[sortedLetters]
	if !ok {
		words[sortedLetters] = []string{}
	}
	// Ignore duplicate words.
	for _, w := range words[sortedLetters] {
		if w == word {
			return
		}
	}
	words[sortedLetters] = append(words[sortedLetters], word)
}

// subtract returns a string containing the letters in the first argument
// minus the letters on the second argument.
func subtract(a, b string) string {
	for ixb := 0; ixb < len(b); ixb++ {
		for ixa := 0; ixa < len(a); ixa++ {
			if a[ixa] == b[ixb] {
				a = a[0:ixa] + a[ixa+1:]
			}
		}
	}
	return a
}

// permutate returns all possible permutations of a single string.
func permutate(prev, s string, res []string) []string {
	ret := res
	//fmt.Printf("[%s] [%s] At entry [%s] [%s]\n", prev, s, res, ret)

	if len(s) != 0 {
		for pos := 0; pos < len(s); pos++ {
			sofar := prev + string(s[pos])
			ret = permutate(sofar, s[pos+1:], ret)
			ret = append(ret, sofar)
		}
	}
	return ret
}

// readDict reads a text fiel containing one word per line into wordLetters.
// Each word is saved in a map keyed by a sorted string containing all letters
// in that word.
func readDict(r io.Reader) (wordLetters, error) {
	words := wordLetters{}
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())
		addUniqueWord(words, sortedString(word, map[rune]bool{'\'': true, ' ': true}), word)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return words, nil
}

// anagram recursively generates anagrams from the passed string and returns
// a slice of anagrams.
func anagram(wl wordLetters, word string, prevwords []string) []string {
	retwords := prevwords
	combos := permutate("", word, []string{})

	for _, combo := range combos {
		// Valid dictionary word?
		anawords, ok := wl[combo]
		if !ok {
			continue
		}

		// Remaining letters in our word minus combo.
		rem := subtract(word, combo)

		// If nothing further to recurse, return.
		if len(rem) == 0 {
			for _, a := range anawords {
				retwords = append(retwords, a)
			}
			continue
		}
		ret := anagram(wl, rem, []string{})

		// Append a all found combinations to the main slice.
		if len(ret) > 0 {
			found := make([]string, len(anawords)*len(ret))
			ix := 0
			for _, a := range anawords {
				for _, w := range ret {
					found[ix] = a + " " + w
					ix++
				}
			}
			retwords = append(retwords, found...)
		}
	}
	return retwords
}
