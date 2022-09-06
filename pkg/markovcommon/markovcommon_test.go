package markovcommon

import (
	"os"
	"path"
	"runtime"
	"testing"

	"golang.org/x/exp/slices"
)

func checkSubSlice[T comparable](s1 []T, s2 []T) bool {
	for _, v := range s2 {
		if !slices.Contains(s1, v) {
			return false
		}
	}
	return true
}

func TestAddStringToData(t *testing.T) {
	testMarkov := MarkovData{}

	// Check valid Data being added
	testMarkov.AddStringToData("This is a test string. I am a test string. My Name is Dr. Rock. Dr. Rock is a professional doctor. Word. Scary stuff! Cool")
	if !checkSubSlice([]string{"This", "I", "My", "Dr.", "Word", "Scary", "Cool"}, testMarkov.Startwords) {
		t.Error("Expected Start words \"This\" and \"I\", got", testMarkov.Startwords, ".")
	}

	// Check nothing being added
	if err := testMarkov.AddStringToData(""); err == nil {
		t.Error("Expected error, did not get one")
	}
}

func TestReadInFile(t *testing.T) {
	if _, err := ReadinFile(""); err == nil {
		t.Error("Expected ReadInFile to throw an error to null filename")
	}

	if _, err := ReadinFile("NotRealFile.txt"); err == nil {
		t.Error("Expected ReadInFile to throw an error for an invalid file")
	}

	if runtime.GOOS == "windows" {
		// try open C:\
		if _, err := ReadinFile("C:\\"); err == nil {
			t.Error("Expected error, got nothing.")
		}
	}

	inp, err := ReadinFile(path.Join("testdata", "test.json"))
	if err != nil {
		t.Error("Could not read valid file.")
	}

	if err := inp.SaveToFile("\n__COM"); err == nil {
		t.Error("This should not have worked")
	}

	if err := inp.SaveToFile(""); err != nil {
		t.Error("Error writing file")
	}

	os.Remove("output.json")

}

func TestCheckValidPath(t *testing.T) {
	if !checkvalidpath("common.go") {
		t.Error("File exists but is not reported.")
	}

	if checkvalidpath("\n__COM") {
		t.Error("Invalid path falsely reported as positive")
	}

	if !checkvalidpath("foo.txt") {
		t.Error("Valid file not writable")
	}
}
