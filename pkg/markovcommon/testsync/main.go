package main

import (
	"fmt"

	"github.com/danielh2942/markov_thingy/pkg/markovcommon"
)

// This allows me to test the markov chain to make sure it doesn't break
func main() {
	md := markovcommon.MarkovData{}

	var opt int
	for {
		fmt.Println("Options:\n0: add sentence\n1: Generate Sentence\n2: Save output\n3: Quit\nChoice: ")
		if _, err := fmt.Scanf("%d", &opt); err != nil {
			fmt.Println("An Error occurred: ", err)
			continue
		}

		switch opt {
		case 0:
			{
				var str string
				fmt.Print("Enter a sentence: ")
				fmt.Scanln(&str)
				if err := md.AddStringToData(str); err != nil {
					fmt.Println("An Error occurred!", err)
				}
			}
		case 1:
			{
				var length int
				fmt.Print("What length do you want? ")
				if _, err := fmt.Scanf("%d", &length); err != nil {
					fmt.Println("An error occurred!", err)
					break
				}

				if txt, err := md.GenerateSentence(length); err == nil {
					fmt.Println(txt)
				} else {
					fmt.Println("An Error Occurred!", err)
				}
			}
		case 2:
			{
				var str string
				fmt.Print("Name the file to save to: ")
				fmt.Scanf("%s", str)
				md.SaveToFile(str)
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
