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

// MarkovCommon
// Author: Daniel Hannon
// Version: 1

type MarkovData struct {
	Startwords []string                  `json:"Startwords"`
	Wordmaps   map[string]map[string]int `json:"Wordmaps"`
}

func checkvalidpath(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}

	if err := os.WriteFile(filename, []byte{}, 0644); err == nil {
		os.Remove(filename)
		return true
	}
	return false
}

// ReadinFile loads a previously saved database file, deserializes it, and returns a MarkovData struct and an error if one were to occur
func ReadinFile(filepath string) (MarkovData, error) {
	if len(filepath) == 0 {
		return MarkovData{}, errors.New("no filename passed, doing nothing")
	}
	if _, err := os.Stat(filepath); err != nil {
		return MarkovData{}, err
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return MarkovData{}, err
	}
	var outp MarkovData
	err1 := json.Unmarshal(data, &outp)
	return outp, err1
}

// SaveToFile exports the current MarkovData struct to a file of choice
// Pass an empty string to save the data to a file called output.json in the current directory
func (md *MarkovData) SaveToFile(filename string) error {
	outpStr, err := json.MarshalIndent(md, "", "\t")
	if err != nil {
		return err
	}

	// Filename verification
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

func checkhonorific(inp string) bool {
	honorifics := []string{"Dr.", "Mrs.", "Ms.", "Prof.", "Rev.", "Sr.", "St."}
	return slices.Contains(honorifics, inp)
}

// AddStringToData parses a string and inserts it in the MarkovData struct as appropriate
func (md *MarkovData) AddStringToData(input string) error {
	if md.Startwords == nil {
		md.Startwords = []string{}
	}
	if md.Wordmaps == nil {
		md.Wordmaps = map[string]map[string]int{}
	}
	// Nothing passed, do nothing.
	if input == "" {
		return errors.New("nothing passed, nothing done")
	}

	// Replace series' of exclamations with one full stop
	exclaimFilter := regexp.MustCompile(`[\!]+`)
	input = exclaimFilter.ReplaceAllString(input, ".")

	generalPuncuationFilter := regexp.MustCompile(`[^a-zA-Z0-9\p{Arabic}\p{Cyrillic}\-.\:<>@_*?']`)
	input = generalPuncuationFilter.ReplaceAllString(input, " ")

	// split on whitespace
	arr := strings.Split(input, " ")

	startOfSentence := true
	var previousWord string
	for _, word := range arr {
		if word == "" {
			continue
		}
		if startOfSentence {
			isStopWord := false
			// Check if a word ends in a full stop rather than contains one as M.D is not a stopword
			if strings.HasSuffix(word, ".") {
				isStopWord = true
				// Check for honorifics/titles
				if checkhonorific(word) {
					isStopWord = false
				} else {
					word = strings.TrimSuffix(word, ".")
				}
			}
			if !slices.Contains(md.Startwords, word) {
				md.Startwords = append(md.Startwords, word)
			}

			if isStopWord {
				if md.Wordmaps[word] == nil {
					md.Wordmaps[word] = map[string]int{}
				}
				md.Wordmaps[word]["\\end"]++
			} else {
				startOfSentence = false
				previousWord = word
			}
			continue
		}
		// Check if it's a stopword
		if strings.HasSuffix(word, ".") && !checkhonorific(word) {
			startOfSentence = true
			word = strings.TrimSuffix(word, ".")
			if md.Wordmaps[previousWord] == nil {
				md.Wordmaps[previousWord] = map[string]int{}
			}
			md.Wordmaps[previousWord][word]++
			if md.Wordmaps[word] == nil {
				md.Wordmaps[word] = map[string]int{}
			}
			md.Wordmaps[word]["\\end"]++
			continue
		}
		if md.Wordmaps[previousWord] == nil {
			md.Wordmaps[previousWord] = map[string]int{}
		}
		md.Wordmaps[previousWord][word]++
		previousWord = word
	}
	if md.Wordmaps[previousWord] == nil {
		md.Wordmaps[previousWord] = map[string]int{}
	}
	if !startOfSentence {
		md.Wordmaps[previousWord]["\\end"]++
	}
	return nil
}

// ReadInTextFile reads in an entire text file and adds to the Markov Chain database
func (md *MarkovData) ReadInTextFile(filename string) error {
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

func weightedpick(inp map[string]int) string {
	tally := 0
	for _, v := range inp {
		tally += v
	}

	choice := rand.Intn(tally) + 1

	threshold := 0
	for k, v := range inp {
		if choice <= (threshold + v) {
			return k
		}
		threshold += v
	}
	return ""
}

func (md *MarkovData) GenerateSentence(limit int) (string, error) {
	if md.Startwords == nil || md.Wordmaps == nil {
		return "", errors.New("no data to generate set is empty")
	}
	currWord := md.Startwords[rand.Intn(len(md.Startwords))]
	output := currWord + " "
	x := 0
	for {
		nextWord := weightedpick(md.Wordmaps[currWord])
		if nextWord == "\\end" || x == limit {
			break
		}
		output += nextWord + " "
		currWord = nextWord
		x++
	}
	return output, nil
}

func (md *MarkovData) Seed() {
	// Seed Random time
	rand.Seed(time.Now().UnixNano())
}
