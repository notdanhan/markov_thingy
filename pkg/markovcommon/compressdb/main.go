package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/danielh2942/markov_thingy/pkg/markovcommon"
)

func main() {
	var input string
	var output string

	flag.StringVar(&input, "input", "", "Input database")
	flag.StringVar(&output, "output", "output.json", "Output database")
	flag.Parse()

	if input == "" {
		fmt.Println("No Input file specified, exiting.")
		return
	}

	content, err := os.ReadFile(input)
	if err != nil {
		fmt.Println("Error occurred", err.Error())
	}

	var InputData markovcommon.MarkovDataOld
	// Create Empty compressed version
	OutputData := markovcommon.MarkovData{}

	if err := json.Unmarshal(content, &InputData); err != nil {
		fmt.Println("Parsing error occurred", err.Error())
		return
	}

	// The actual data transformation step
	// Step 1 create the List of WordRef and WordVals
	OutputData.WordCount = 0
	OutputData.WordRef = map[string]uint{}
	OutputData.WordVals = []string{}
	for k := range InputData.Wordmaps {
		if _, ok := OutputData.WordRef[k]; !ok {
			OutputData.WordRef[k] = OutputData.WordCount
			OutputData.WordVals = append(OutputData.WordVals, k)
			OutputData.WordCount++
		}
	}
	OutputData.WordRef["."] = OutputData.WordCount
	OutputData.WordVals = append(OutputData.WordVals, ".")
	OutputData.WordCount++

	OutputData.WordGraph = []map[uint]uint{}
	// Now the actual mappings
	for idx, val := range OutputData.WordVals {
		OutputData.WordGraph = append(OutputData.WordGraph, map[uint]uint{})
		for k, v := range InputData.Wordmaps[val] {
			if k == "\\end" {
				OutputData.WordGraph[idx][OutputData.WordRef["."]] = uint(v)
			} else {
				OutputData.WordGraph[idx][OutputData.WordRef[k]] = uint(v)
			}
		}
	}

	OutputData.StartWords = []uint{}
	for _, v := range InputData.Startwords {
		OutputData.StartWords = append(OutputData.StartWords, OutputData.WordRef[v])
	}

	OutputData.SaveToFile(output)
}
