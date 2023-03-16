 # Markov_thingy

markov_thingy is a markov-chain discord bot maintained primarily by danielh2942
It allows you to train a markov chain off a given channel in your discord server.
It was written for fun and is probably going to be maintained on a very ad-hoc basis

## Features

- Train a Markov bot off of discord messages sent to a channel
- Weird youtube searches using the youtube API
- Multiple server support

## Tech

Dillinger uses a number of open source projects to work properly:

- [Go](https://go.dev) - The Go programming language
- [DiscordGo](https://github.com/bwmarrin/discordgo) - A discord library for the go programming language
- [google/uuid](https://github.com/google/uuid) - Googles UUID library written in go, 

And of course markov_thingy itself is open source.

## Installation

markov_thingy requires [Go](https://go.dev/) 1.19 to run.

Install the dependencies

```sh
go mod tidy
```

Write your config file as such
```json
{
    "Token":"Your Discord Token",
    "YoutubeAPIKey":"Your youtube API key",
    "Prefix":"Your prefix of choice",
    "Servers":{}
}
```
save it as config.json in the same directory as the executable and run!

## Development

Want to contribute? Great!

Please submit patches to the dev branch with anything that you think would contribute nicely to the bot :)

Please attach the pre-commit hook in .githooks before you add anything to the repo!
(It uses gofmt so install that too.)

## Contact
If you need to get in touch with me contact me on discord or by email
Discord: notdanhan#3199
Email: [danhan@live.ie](mailto:danhan@live.ie)

Thanks and have fun!
