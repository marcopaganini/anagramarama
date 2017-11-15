// Unit tests for anagramarama.

package main

import (
	"sort"
	"testing"
)

func TestAnagram(t *testing.T) {
	casetests := []struct {
		phrase    string
		dictFile  string
		wantFile  string
		parallel  int
		wantError bool
	}{
		// One thread.
		{
			phrase:    "marco paganini ab",
			dictFile:  "testdata/words.txt",
			wantFile:  "testdata/results.txt",
			parallel:  1,
			wantError: false,
		},
		// Multiple threads.
		{
			phrase:    "marco paganini ab",
			dictFile:  "testdata/words.txt",
			wantFile:  "testdata/results.txt",
			parallel:  16,
			wantError: false,
		},
		// Invalid dictionary name (error)
		{
			phrase:    "marco paganini ab",
			dictFile:  "INVALIDFILE",
			wantFile:  "testdata/results.txt",
			parallel:  1,
			wantError: true,
		},
	}

	for _, tt := range casetests {
		phrase, err := sanitize(tt.phrase)
		if err != nil {
			t.Fatalf("error sanitizing phrase: %v", err)
		}

		// Results file
		want, err := readDict(tt.wantFile)
		if err != nil {
			t.Fatalf("error reading results file: %v", err)
		}

		words, err := readDict(tt.dictFile)
		if !tt.wantError {
			if err != nil {
				t.Fatalf("Got error %q want no error", err)
			}

			// Generate list of candidate and alternate words.
			cand, altwords := candidates(words, phrase)
			got := anagrams(phrase, cand, altwords, tt.parallel)

			lenGot := len(got)
			lenWant := len(want)
			if lenGot != lenWant {
				t.Fatalf("anagram lists have different lengths. Got %d lines, want %d lines.", lenGot, lenWant)
			}

			// Sort output by words and then by line.
			sortWords(got)
			sort.Strings(got)

			for ix := 0; ix < lenGot; ix++ {
				if got[ix] != want[ix] {
					t.Fatalf("diff: line %d, Got %q, want %q.", ix, got[ix], want[ix])
				}
			}
			continue
		}

		// Here, we want to see an error.
		if err == nil {
			t.Errorf("Got no error, want error")
		}
	}
}