package main

import (
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/danielh2942/markov_thingy/pkg/markovcommon"
	"github.com/danielh2942/markov_thingy/pkg/youtubesearch"
)

type AuthStruct struct {
	Token         string `json:"Token"`         // Discord Auth Token
	YoutubeAPIKey string `json:"YoutubeAPIKey"` // Youtube Data Api Key token
	Prefix        string `json:"Prefix"`        // Command Prefix (TODO: Remove in favor of slash commands)
	Lock          bool   `json:"ChanLock"`      // Locked to channel
	ChanId        string `json:"ChanId"`        // Channel locked to
}

type ProgramFlags struct {
	Save        bool   // Save database incrementally
	LogToFile   bool   // Write logs to a file (enforced form markov_bot_[date]_log.txt)
	InputData   string // Preexisting dataset
	PostingOdds uint   // Odds out of 100 that it will reply
	BackupFreq  uint   // Save backup every n messages
}

func (pf ProgramFlags) String() string {
	var output string
	output += "Program Flags\n"
	output += "Save database:\t\t\t" + strconv.FormatBool(pf.Save) + "\n"
	output += "Save Logs as file:\t\t\t" + strconv.FormatBool(pf.LogToFile) + "\n"
	output += "Input Data File:\t\t\t" + pf.InputData + "\n"
	output += "Response Frequency:\t\t\t" + strconv.FormatUint(uint64(pf.PostingOdds), 10) + "/100\n"
	output += "Save Messages Every " + strconv.FormatUint(uint64(pf.BackupFreq), 10) + " Messages\n"
	return output
}

func GetFlags() ProgramFlags {
	progFlags := ProgramFlags{}

	flag.BoolVar(&progFlags.Save, "save", false, "Save the data provided")
	flag.StringVar(&progFlags.InputData, "inp", "", "The database to work off")
	flag.UintVar(&progFlags.PostingOdds, "odds", 20, "Likelihood out of 100")
	flag.BoolVar(&progFlags.LogToFile, "savelogs", false, "Log to a file")
	flag.UintVar(&progFlags.BackupFreq, "backup", 100, "How many messages before a backup")

	flag.Parse()

	return progFlags
}

func main() {

	var logger *log.Logger
	var file *os.File = nil

	progFlags := GetFlags()

	if progFlags.LogToFile {
		file, err := os.Create(("markov_bot_" + strings.ReplaceAll(time.Now().Local().Format(time.RFC3339), ":", "_") + "_log.txt"))
		if err != nil {
			log.Fatalln("FATAL ERROR:", err.Error())
		}
		logger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	} else {
		logger = log.Default()
	}

	logger.Println(progFlags)

	if file != nil {
		defer file.Close()
	}

	var markov markovcommon.MarkovData
	var err error
	if progFlags.InputData != "" {
		logger.Println("Loading in database:", progFlags.InputData)
		markov, err = markovcommon.ReadinFile(progFlags.InputData)
		if err != nil {
			logger.Fatalln("FATAL ERROR:", err.Error())
		}
		logger.Println("Done")
	} else {
		logger.Println("No database passed, creating empty database")
		markov = markovcommon.MarkovData{}
		progFlags.InputData = "db.json"
	}

	logger.Println("Reading in config file")
	inpFile, err := os.ReadFile("config.json")
	if err != nil {
		logger.Fatalln("FATAL ERROR", err.Error())
	}

	var myAuth AuthStruct
	err = json.Unmarshal(inpFile, &myAuth)
	if err != nil {
		logger.Fatalln("FATAL ERROR: Failed to read config.json. Reason:", err.Error())
	}
	discbot, err := discordgo.New("Bot " + myAuth.Token)
	if err != nil {
		logger.Fatalln(err.Error())
	}

	// get Bot id
	u, err := discbot.User("@me")
	if err != nil {
		logger.Fatalln("FATAL ERROR", err.Error())
	}
	BotId := u.ID
	count := progFlags.BackupFreq
	logger.Println("Setting up Youtube API stuff")
	ytListener := youtubesearch.New(myAuth.YoutubeAPIKey, logger)
	defer ytListener.Close()
	logger.Println("Connecting general operation loop")
	discbot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == BotId {
			return
		}
		if strings.HasPrefix(m.Content, myAuth.Prefix) {
			if m.Message.Content == myAuth.Prefix+"bark" {
				msg, err := markov.GenerateSentence(50)
				if err != nil {
					logger.Println("Non-Fatal Error:", err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, msg)
			}
			if m.Message.Content == myAuth.Prefix+"lock" {
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
				logger.Println("Messages are now only read from channel with ID", m.ChannelID)
				return
			}
			if m.Message.Content == myAuth.Prefix+"save" && progFlags.Save {
				// force save - this is for debugging
				if err := markov.SaveToFile(progFlags.InputData); err != nil {
					logger.Println("Non-Fatal Error:", err.Error())
				} else {
					logger.Println("Saving checkpoint.")
				}
				return
			}
			if strings.HasPrefix(m.Message.Content, myAuth.Prefix+"setbackup") && progFlags.Save {
				numFilter := regexp.MustCompile(`[0-9]+`)
				val, err := strconv.Atoi(numFilter.FindString(m.Message.Content))
				if err != nil {
					logger.Println("Non-Fatal error:", err.Error())
					return
				}
				progFlags.BackupFreq = uint(val)
				logger.Println("Backup frequency changed to every ", val, "Messages!")
				return
			}
			if strings.HasPrefix(m.Message.Content, myAuth.Prefix+"adjustrate") {
				numFilter := regexp.MustCompile(`[0-9]+`)
				val, err := strconv.Atoi(numFilter.FindString(m.Message.Content))
				if err != nil {
					logger.Println("Non-Fatal error:", err.Error())
					return
				}
				if val > 100 {
					logger.Println("Invalid number entered for rate", val)
					return
				}
				progFlags.PostingOdds = uint(val)
			}
			if m.Message.Content == myAuth.Prefix+"ytrandom" && m.ChannelID == myAuth.ChanId {
				mq, err := markov.GenerateSentence(20)
				if err != nil {
					logger.Println("Failed to generate Sentence, reason:", err.Error())
				}
				s.ChannelMessageSend(m.ChannelID, "Video found with Query \""+mq+"\"\n"+ytListener.GetRandomVid(mq))
				return
			}
			if m.Message.Content == myAuth.Prefix+"help" && m.ChannelID == myAuth.ChanId {
				s.ChannelMessageSend(m.ChannelID, "```"+myAuth.Prefix+"help\t\t\tShows this\n"+myAuth.Prefix+"ytrandom\t\tRandom Youtube Video from search query generated from input data\n"+myAuth.Prefix+"bark\t\t\tSay Something\n"+myAuth.Prefix+"adjustrate <value 0-100>\t\tChances out of 100 that the bot will say something```")
				return
			}
		} else {
			if myAuth.Lock && m.ChannelID != myAuth.ChanId {
				return
			}
			markov.AddStringToData(m.Content)
			// save in bursts of n messages
			if progFlags.Save && count >= progFlags.BackupFreq {
				if err := markov.SaveToFile(progFlags.InputData); err != nil {
					logger.Println("Non-Fatal Error", err.Error())
				} else {
					logger.Println("Saving checkpoint.")
				}
				count = 0
			}
			// Reply when mentioned
			if len(m.Mentions) > 0 {
				for _, ment := range m.Mentions {
					if ment.ID == BotId {
						msg, err := markov.GenerateSentence(50)
						if err != nil {
							logger.Println("Non-fatal ERROR:", err.Error())
						}
						s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
					}
				}
			} else if rand.Intn(100) < int(progFlags.PostingOdds) {
				msg, err := markov.GenerateSentence(50)
				if err != nil {
					logger.Println("Non-fatal ERROR:", err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, msg)
			}
			count++
		}
	})

	// Only care about messages
	discbot.Identify.Intents |= discordgo.IntentsGuildMessages

	logger.Println("Initalizing Discord Bot")

	err = discbot.Open()
	if err != nil {
		logger.Fatalln("FATAL ERROR:", err.Error())
	}
	logger.Println("Bot Initalized")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sc

	// Cleanly close down the Discord session.
	discbot.Close()
	// Save whatever the hell it had at the time of shutdown
	logger.Println("Shutting down.")
	if progFlags.Save {
		markov.SaveToFile(progFlags.InputData)
	}
}
