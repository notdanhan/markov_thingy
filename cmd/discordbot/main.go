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
	"github.com/danielh2942/markov_thingy/pkg/servsync"
	"github.com/danielh2942/markov_thingy/pkg/youtubesearch"
)

type AuthStruct struct {
	Token         string           `json:"Token"`         // Discord Auth Token
	YoutubeAPIKey string           `json:"YoutubeAPIKey"` // Youtube Data Api Key token
	Prefix        string           `json:"Prefix"`        // Command Prefix (TODO: Remove in favor of slash commands)
	Servers       servsync.SyncMap `json:"Servers"`       // The servers that the program has access to
}

type ProgramFlags struct {
	Save        bool   // Save database incrementally
	LogToFile   bool   // Write logs to a file (enforced form markov_bot_[date]_log.txt)
	PostingOdds uint   // Odds out of 100 that it will reply
	BackupFreq  uint64 // Save backup every n messages
}

func (pf ProgramFlags) String() string {
	var output string
	output += "\n____Program Flags____\n"
	output += "Save database:\t\t" + strconv.FormatBool(pf.Save) + "\n"
	output += "Save Logs as file:\t" + strconv.FormatBool(pf.LogToFile) + "\n"
	output += "Response Frequency:\t" + strconv.FormatUint(uint64(pf.PostingOdds), 10) + "/100\n"
	output += "Save Messages Every " + strconv.FormatUint(uint64(pf.BackupFreq), 10) + " Messages\n"
	return output
}

func GetFlags() ProgramFlags {
	progFlags := ProgramFlags{}

	flag.BoolVar(&progFlags.Save, "nosave", true, "Don't save the data provided")
	flag.UintVar(&progFlags.PostingOdds, "odds", 20, "Likelihood out of 100")
	flag.BoolVar(&progFlags.LogToFile, "savelogs", false, "Log to a file")
	flag.Uint64Var(&progFlags.BackupFreq, "backup", 100, "How many messages before a backup")

	flag.Parse()

	return progFlags
}

var (
	progFlags             = GetFlags()
	logger    *log.Logger = nil
	file      *os.File    = nil
	BotId     string
)

func main() {
	if progFlags.LogToFile {
		file, err := os.Create(
			("markov_bot_" + strings.ReplaceAll(time.Now().Local().Format(time.RFC3339), ":", "_") + "_log.txt"),
		)
		if err != nil {
			log.Fatalln("FATAL ERROR:", err.Error())
		}
		logger = log.New(file, "", log.LstdFlags|log.Lmicroseconds|log.Lmsgprefix|log.Lshortfile)
		// Set general logs to output to file too (this is stuff related to third party libs)
		log.SetOutput(file)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lmsgprefix | log.Lshortfile)
	} else {
		logger = log.Default()
	}
	logger.SetPrefix("[Markov Discord Bot] ")

	logger.Println(progFlags)

	if file != nil {
		defer file.Close()
	}

	var err error
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
	BotId = u.ID
	logger.Println("Setting up Youtube API stuff")
	ytListener := youtubesearch.New(myAuth.YoutubeAPIKey, logger)
	defer ytListener.Close()
	logger.Println("Connecting general operation loop")
	discbot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		serv, exists := myAuth.Servers.Get(m.GuildID)
		if m.Author.ID == BotId {
			return
		}
		if strings.HasPrefix(m.Content, myAuth.Prefix) {
			if m.Message.Content == myAuth.Prefix+"bark" {
				if serv == nil {
					return
				}
				msg, err := serv.MarkovChain.GenerateSentence(50)
				if err != nil {
					logger.Println("Non-fatal Error:", err.Error())
					return
				}
				s.ChannelMessageSend(serv.ChanId, msg)
			}
			if m.Message.Content == myAuth.Prefix+"lock" {
				// limit to one channel
				if !exists {
					mc := servsync.New(m.ChannelID)
					myAuth.Servers.Set(m.GuildID, mc)
				} else {
					serv.ChanId = m.ChannelID
				}
				outp, err := json.MarshalIndent(myAuth, "", "\t")
				if err != nil {
					return
				}
				outFile, err := os.Create("config.json")
				if err != nil {
					return
				}
				outFile.Write(outp)
				outFile.Close()
				logger.Println("Messages from guild", m.GuildID, "are now only read from channel with ID", m.ChannelID)
				return
			}
			if m.Message.Content == myAuth.Prefix+"save" && progFlags.Save {
				// force save - this is for debugging
				if !exists {
					return
				}
				if err := serv.Save(); err != nil {
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
				progFlags.BackupFreq = uint64(val)
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
			if !exists {
				return
			}
			if m.Message.Content == myAuth.Prefix+"ytrandom" && m.ChannelID == serv.ChanId {
				// Make sure that it always returns a video
				for {
					mq, err := serv.MarkovChain.GenerateSentence(20)
					if err != nil {
						logger.Println("Failed to generate Sentence, reason:", err.Error())
						s.ChannelMessageSend(m.ChannelID, "An Error Occurred while generating a sentence!")
						return
					}
					vid, err := ytListener.GetRandomVid(mq)
					if err == nil {
						s.ChannelMessageSend(m.ChannelID, "Video found with Query \""+mq+"\"\n"+vid)
						break
					}
				}
				return
			}
			if m.Message.Content == myAuth.Prefix+"help" && m.ChannelID == serv.ChanId {
				s.ChannelMessageSend(
					m.ChannelID,
					"```"+myAuth.Prefix+"help\t\t\tShows this\n"+myAuth.Prefix+"ytrandom\t\tRandom Youtube Video from search query generated from input data\n"+myAuth.Prefix+"bark\t\t\tSay Something\n"+myAuth.Prefix+"adjustrate <value 0-100>\t\tChances out of 100 that the bot will say something```",
				)
				return
			}
		} else {
			if m.ChannelID != serv.ChanId {
				return
			}
			serv.MarkovChain.AddStringToData(m.Content)
			// save in bursts of n messages
			if progFlags.Save && serv.MsgCount.Load() >= progFlags.BackupFreq {
				if err := serv.Save(); err != nil {
					logger.Println("Non-Fatal Error", err.Error())
				} else {
					logger.Println("Saving checkpoint.")
				}
				serv.MsgCount.Store(0)
			}
			// Reply when mentioned
			if len(m.Mentions) > 0 {
				for _, ment := range m.Mentions {
					if ment.ID == BotId {
						msg, err := serv.MarkovChain.GenerateSentence(50)
						if err != nil {
							logger.Println("Non-fatal ERROR:", err.Error())
						}
						s.ChannelMessageSendReply(m.ChannelID, msg, m.Reference())
					}
				}
			} else if rand.Intn(100) < int(progFlags.PostingOdds) {
				msg, err := serv.MarkovChain.GenerateSentence(50)
				if err != nil {
					logger.Println("Non-fatal ERROR:", err.Error())
					return
				}
				s.ChannelMessageSend(m.ChannelID, msg)
			}
			serv.MsgCount.Add(1)
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
		outp, err := json.MarshalIndent(myAuth, "", "\t")
		if err != nil {
			return
		}
		outFile, err := os.Create("config.json")
		if err != nil {
			return
		}
		outFile.Write(outp)
		outFile.Close()
	}
}
