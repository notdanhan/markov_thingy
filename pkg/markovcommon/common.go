package markovcommon

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
)

// MarkovCommon
// Author: Daniel Hannon
// Version: 1

type MarkovChain interface {
	AddStringToData(string) error
	ReadInTextFile(string) error
	GenerateSentence(int) (string, error)
	SaveToFile(string) error
}

// Helper functions

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

// ReadinFile loads a previously saved database file, deserializes it, and returns a struct matching the MarkovChain interface
func ReadinFile(filepath string) (MarkovChain, error) {
	if len(filepath) == 0 || filepath == "" {
		return &MarkovData{}, errors.New("no filename passed, doing nothing")
	}
	if _, err := os.Stat(filepath); err != nil {
		return &MarkovData{}, err
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return &MarkovData{}, err
	}
	var outp MarkovData
	err1 := json.Unmarshal(data, &outp)
	if err1 == nil {
		return &outp, nil
	}
	var outp1 MarkovDataOld
	err1 = json.Unmarshal(data, &outp1)
	return &outp1, err1
}

func checkhonorific(inp string) bool {
	honorifics := []string{"Dr.", "Mrs.", "Ms.", "Prof.", "Rev.", "Sr.", "St."}
	return slices.Contains(honorifics, inp)
}
