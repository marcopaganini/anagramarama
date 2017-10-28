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
	"log"
	"os"
	"runtime/pprof"
)

func main() {
	var (
		optDict       string
		optCPUProfile string
	)

	log.SetFlags(0)

	flag.StringVar(&optDict, "dict", "/usr/share/dict/words", "dictionary file")
	flag.StringVar(&optCPUProfile, "cpuprofile", "", "write cpu profile to file")
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

	file, err := os.Open(optDict)
	if err != nil {
		log.Fatalln(err)
	}
	wlist, err := readDict(file)
	if err != nil {
		log.Fatalln(err)
	}

	word := sortedString(flag.Args()[0], map[rune]bool{})

	res := anagram(wlist, word, []string{})
	for _, s := range res {
		fmt.Println(s)
	}
}
