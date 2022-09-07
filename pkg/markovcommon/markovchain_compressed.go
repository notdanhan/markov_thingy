package markovcommon

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

// MarkovDataAlt
// Author: Daniel Hannon
// Version: 1
// Brief: This is like MarkovData but it uses some compression shit innit

type MarkovDataAlt struct {
	StartWords []uint          `json:"StartWords"` // Numeric references to each start word
	WordCount  uint            `json:"WordCount"`  // Number of words available
	WordRef    map[string]uint `json:"WordMap"`    // Word to number mappings
	WordVals   []string        `json:"WordVals"`   // Number to word mappings
	WordGraph  []map[uint]uint `json:"WordGraph"`  // Mappings of word number -> word number with frequency of relationship
}

// getWordRef checks if a word exists and returns it's numeric equivalent, otherwise it makes one :)
func (md *MarkovDataAlt) getWordRef(word string) uint {
	if v, ok := md.WordRef[word]; ok {
		return v
	}
	md.WordRef[word] = md.WordCount
	md.WordVals = append(md.WordVals, word)
	temp := md.WordCount
	md.WordGraph = append(md.WordGraph, map[uint]uint{})
	md.WordCount++
	return temp
}

// AddStringToData gets a string and parses it into a format that is interpretable by the MarkovData struct
func (md *MarkovDataAlt) AddStringToData(input string) error {
	if input == "" {
		return errors.New("nothing passed, nothing to do")
	}

	// Initialization checks
	// This makes sure nothing fucks up
	if md.StartWords == nil {
		md.StartWords = []uint{}
	}
	if md.WordRef == nil {
		md.WordRef = map[string]uint{}
	}
	if md.WordVals == nil {
		md.WordVals = []string{}
		md.WordCount = uint(len(md.WordVals))
	}
	if md.WordGraph == nil {
		md.WordGraph = []map[uint]uint{}
	}

	// Some Sanitization for reasons

	// Filter out illegal characters
	generalPuncuationFilter := regexp.MustCompile(`[^a-zA-Z0-9\p{Arabic}\p{Cyrillic}\-.\:\/\\!,.<>@_*?']`)
	input = generalPuncuationFilter.ReplaceAllString(input, "")

	// Separate exclamations
	exclaimFilter := regexp.MustCompile(`[^a-zA-Z0-9\p{Arabic}\p{Cyrillic}]+[!]+`)
	input = exclaimFilter.ReplaceAllStringFunc(input, func(inp string) string {
		exclaimFilter1 := regexp.MustCompile(`[!]+`)
		return exclaimFilter1.ReplaceAllString(inp, " ! ")
	})

	// Separate commas
	commaFilter := regexp.MustCompile(`[a-zA-Z0-9\p{Arabic}\p{Cyrillic}]+,`)
	input = commaFilter.ReplaceAllStringFunc(input, func(inp string) string {
		commaFilter1 := regexp.MustCompile(`[,]+`)
		return commaFilter1.ReplaceAllString(inp, " , ")
	})

	// Separate Full Stops
	fullStopFilter := regexp.MustCompile(`[a-zA-Z0-9\p{Arabic}\p{Cyrillic}]+\.\s`)
	input = fullStopFilter.ReplaceAllStringFunc(input, func(inp string) string {
		if checkhonorific(inp) {
			return inp
		}
		fullStopFilter1 := regexp.MustCompile(`\.\s`)
		return fullStopFilter1.ReplaceAllString(inp, " . ")
	})

	// Split input into tokens
	wordArr := strings.Split(input, " ")
	startOfSentence := true
	var previousWord uint

	// Insert the data as appropriate
	for _, word := range wordArr {
		if len(word) == 0 {
			continue
		}
		if startOfSentence {
			if strings.Contains(".,!", word) {
				continue
			}

			startOfSentence = false
			val := md.getWordRef(word)
			if !slices.Contains(md.StartWords, val) {
				md.StartWords = append(md.StartWords, val)
			}
			previousWord = val
			continue
		}
		currWord := md.getWordRef(word)
		md.WordGraph[previousWord][currWord]++
		previousWord = currWord

		// Check stopwords
		if strings.Contains(".!", word) {
			startOfSentence = true
		}
	}

	// Don't add data to stop words, no point.
	if previousWord == md.getWordRef(".") || previousWord == md.getWordRef("!") {
		return nil
	} else {
		md.WordGraph[previousWord][md.getWordRef(".")]++
	}
	return nil
}

func (md *MarkovDataAlt) weightedPick(wordNo uint) uint {
	tally := 0
	for _, v := range md.WordGraph[wordNo] {
		tally += int(v)
	}

	choice := rand.Intn(tally + 1)
	offset := 0
	for k, v := range md.WordGraph[wordNo] {
		offset += int(v)
		if choice <= offset {
			return k
		}
	}
	return md.getWordRef(".")
}

// ReadInTextFile reads in an entire text file and adds to the Markov Chain database
func (md *MarkovDataAlt) ReadInTextFile(filename string) error {
	if !checkvalidpath(filename) {
		return errors.New("path of text file is invalid")
	}
	inp, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	inpStr := string(inp)
	inpStr = strings.ReplaceAll(inpStr, "\n", "")
	return md.AddStringToData(inpStr)
}

// GenerateSentence produces a sentence using the provided database
func (md *MarkovDataAlt) GenerateSentence(limit int) (string, error) {
	if md.WordCount == 0 {
		return "", errors.New("no data in markov database")
	}
	currWord := md.StartWords[rand.Intn(len(md.StartWords))]
	output := md.WordVals[currWord]
	x := 0
	for x < limit {
		nextWord := md.weightedPick(currWord)
		if strings.Contains(".!", md.WordVals[currWord]) {
			output += md.WordVals[nextWord]
			break
		}
		output += " " + md.WordVals[nextWord]
		currWord = nextWord
		x++
	}
	return output, nil
}

// SaveToFile outputs the data generated to a file, since it's not exactly human readable, it's just clumped together
func (md *MarkovDataAlt) SaveToFile(filename string) error {
	outpStr, err := json.Marshal(md)
	if err != nil {
		return err
	}

	// Verify filename
	if filename == "" {
		// Default export
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		filename = path.Join(wd, "output.json")
	} else if !checkvalidpath(filename) {
		return errors.New("invalid file path provided")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(outpStr)
	return nil
}

// Seed seeds the RNG for the markov num gen
func (md *MarkovDataAlt) Seed() {
	//Seed random time
	rand.Seed(time.Now().UnixNano())
}
