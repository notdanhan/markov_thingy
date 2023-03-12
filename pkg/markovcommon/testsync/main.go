package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/danielh2942/markov_thingy/pkg/markovcommon"
)

// This allows me to test the markov chain to make sure it doesn't break
func main() {
	md := markovcommon.MarkovData{}

	scanner := bufio.NewReader(os.Stdin)

	var opt int
	for {
		fmt.Print("Options:\n0: add sentence\n1: Generate Sentence\n2: Save output\n3: Quit\nChoice: ")
		if optio, _, err := scanner.ReadLine(); err != nil {
			fmt.Println("An Error occurred: ", err)
			continue
		} else {
			if x, err := strconv.Atoi(string(optio)); err != nil {
				fmt.Println(err)
				continue
			} else {
				opt = x
			}
		}

		switch opt {
		case 0:
			{
				fmt.Print("Enter a sentence: ")
				str, _, _ := scanner.ReadLine()
				if err := md.AddStringToData(string(str)); err != nil {
					fmt.Println("An Error occurred!", err)
				}
			}
		case 1:
			{
				var length int
				fmt.Print("What length do you want? ")
				if length1, _, err := scanner.ReadLine(); err != nil {
					fmt.Println("An error occurred!", err)
					break
				} else {
					if x, err := strconv.Atoi(string(length1)); err != nil {
						fmt.Println(err)
						break
					} else {
						length = x
					}
				}

				if txt, err := md.GenerateSentence(length); err == nil {
					fmt.Println(txt)
				} else {
					fmt.Println("An Error Occurred!", err)
				}
			}
		case 2:
			{
				fmt.Print("Name the file to save to: ")
				str, _, _ := scanner.ReadLine()
				md.SaveToFile(string(str))
			}
		case 3:
			{
				return
			}
		default:
			{
				fmt.Println("Invalid choice.")
			}
		}
	}
}
