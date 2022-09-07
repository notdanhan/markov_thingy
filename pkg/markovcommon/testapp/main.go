package main

import (
	"flag"
	"fmt"

	"github.com/danielh2942/markov_thingy/pkg/markovcommon"
)

// markovcommon testapp
// this reads in files passed and builds a markov db accordingly.

func main() {
	var database string
	var inputFile string

	flag.StringVar(&database, "data", "", "Markov Database that exists (none by default)")
	flag.StringVar(&inputFile, "inp", "", "File to use to extend the database.")

	flag.Parse()

	if inputFile == "" {
		fmt.Println("No data file passed, nothing to do.")
		return
	}

	var myMarkov markovcommon.MarkovChain
	var err error
	if database == "" {
		myMarkov = &markovcommon.MarkovData{}
	} else {
		if myMarkov, err = markovcommon.ReadinFile(database); err != nil {
			fmt.Println("Error Occurred", err.Error())
			return
		}
	}

	myMarkov.ReadInTextFile(inputFile)
	myMarkov.SaveToFile(database)
	myMarkov.Seed()
	for i := 0; i < 10; i++ {
		fmt.Println(myMarkov.GenerateSentence(999))
	}
}
