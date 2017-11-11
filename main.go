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

func main() {
	var (
		optDict       string
		optCPUProfile string
		optSilent     bool
		optCandidates bool
	)

	log.SetFlags(0)

	flag.StringVar(&optDict, "dict", "words.txt", "dictionary file")
	flag.StringVar(&optCPUProfile, "cpuprofile", "", "write cpu profile to file")
	flag.BoolVar(&optSilent, "silent", false, "don't print results.")
	flag.BoolVar(&optCandidates, "candidates", false, "just show candidate words (don't anagram)")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Use: anagramarama word")
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

	buf, err := ioutil.ReadFile(optDict)
	if err != nil {
		log.Fatalln(err)
	}

	// Convert expression to uppercase
	// FIXME: Optimize this.
	phrase, err := sanitize(flag.Args()[0])
	if err != nil {
		log.Fatalln(err)
	}

	words := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	cand := candidates(words, phrase)

	if optCandidates {
		for _, w := range cand {
			fmt.Println(w)
		}
		os.Exit(0)
	}

	an := anagrams(phrase, cand)

	if !optSilent {
		for _, w := range an {
			fmt.Println(w)
		}
	}
}
