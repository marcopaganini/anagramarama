// This file is part of anagramarama - A simple anagram generator in Go.
//
// (C) Oct/2017 by Marco Paganini <paganini@paganini.net>
//
// Check the main repository at http://github.com/marcopaganini/anagramarama
// for more details.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime/pprof"
	"sort"
	"strings"
)

// Sanitize converts the input string to uppercase and removes all characters
// that don't match [A-Z].
func sanitize(s string) (string, error) {
	re, err := regexp.Compile("[^A-Z]")
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(strings.ToUpper(s), ""), nil
}

// sortWords reads a slice of strings and sorts each line by word.
func sortWords(lines []string) {
	for idx, line := range lines {
		w := strings.Split(line, " ")
		sort.Strings(w)
		lines[idx] = strings.Join(w, " ")
	}
}

func main() {
	var (
		optCandidates bool
		optCPUProfile string
		optDict       string
		optParallel   int
		optSilent     bool
		optSortLines  bool
		optSortWords  bool
	)

	log.SetFlags(0)

	flag.BoolVar(&optCandidates, "candidates", false, "just show candidate words (don't anagram)")
	flag.StringVar(&optCPUProfile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&optDict, "dict", "words.txt", "dictionary file")
	flag.IntVar(&optParallel, "parallelism", 16, "number of goroutine threads")
	flag.BoolVar(&optSilent, "silent", false, "don't print results.")
	flag.BoolVar(&optSortLines, "sortlines", false, "(also) sort the output by lines")
	flag.BoolVar(&optSortWords, "sortwords", true, "(also) sort the output by words")

	// Custom usage.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Use: %s [flags] expression_to_anagram\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(2)
	}
	// Profiling
	if optCPUProfile != "" {
		f, err := os.Create(optCPUProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Read entire file in memory.
	buf, err := ioutil.ReadFile(optDict)
	if err != nil {
		log.Fatalln(err)
	}

	phrase, err := sanitize(flag.Args()[0])
	if err != nil {
		log.Fatalln(err)
	}

	// Split input in newlines and generate the list of candidate words.
	words := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	cand, altwords := candidates(words, phrase)

	if optCandidates {
		for _, w := range cand {
			fmt.Println(w)
		}
		os.Exit(0)
	}

	// Anagram & Print sorted by word (and optionally, by line.)
	an := anagrams(phrase, cand, altwords, optParallel)

	if !optSilent {
		if optSortWords {
			sortWords(an)
		}
		if optSortLines {
			sort.Strings(an)
		}
		for _, w := range an {
			fmt.Println(w)
		}
	}
}
