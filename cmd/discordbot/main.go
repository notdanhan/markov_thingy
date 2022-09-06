package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/danielh2942/markov_thingy/pkg/markovcommon"
)

type AuthStruct struct {
	Token  string `json:"Token"`
	Prefix string `json:"Prefix"`
	Lock   bool   `json:"ChanLock"`
	ChanId string `json:"ChanId"`
}

func main() {

	var save bool
	var inputData string
	var odds int

	flag.BoolVar(&save, "save", false, "Save the data provided")
	flag.StringVar(&inputData, "inp", "", "The database to work off")
	flag.IntVar(&odds, "odds", 20, "Likelihood out of 100")

	flag.Parse()

	var markov markovcommon.MarkovData
	var err error
	if inputData != "" {
		markov, err = markovcommon.ReadinFile(inputData)
		if err != nil {
			fmt.Println("ERROR:", err.Error())
			return
		}
	} else {
		markov = markovcommon.MarkovData{}
		inputData = "db.json"
	}

	inpFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("ERROR", err.Error())
		return
	}

	var myAuth AuthStruct
	json.Unmarshal(inpFile, &myAuth)

	discbot, err := discordgo.New("Bot " + myAuth.Token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// get Bot id
	u, err := discbot.User("@me")
	if err != nil {
		fmt.Println("ERROR COULD NOT DO THING", err.Error())
		return
	}
	BotId := u.ID
	count := 100

	discbot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotId {
			return
		}
		if strings.HasPrefix(m.Content, myAuth.Prefix) {
			if m.Message.Content == "!bark" {
				msg, err := markov.GenerateSentence(50)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, msg)
			}
			if m.Message.Content == "!lock" {
				// limit to one channel
				myAuth.Lock = true
				myAuth.ChanId = m.ChannelID
				outp, err := json.MarshalIndent(myAuth, "", "\t")
				if err != nil {
					return
				}
				outFile, err := os.Create("config.json")
				if err != nil {
					return
				}
				defer outFile.Close()
				outFile.Write(outp)
				return
			}
		} else {
			if myAuth.Lock && m.ChannelID != myAuth.ChanId {
				return
			}
			markov.AddStringToData(m.Content)
			// save every 100 messages
			if save && count == 100 {
				markov.SaveToFile(inputData)
				count = 0
			}
			if rand.Intn(100) < odds {
				msg, err := markov.GenerateSentence(50)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, msg)
			}
			count++
		}
	})

	// Only care about messages
	discbot.Identify.Intents |= discordgo.IntentsGuildMessages

	err = discbot.Open()
	if err != nil {
		fmt.Println("ERROR:", err.Error())
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	discbot.Close()
}
