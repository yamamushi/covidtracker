package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

type CommandParser struct {
	dg *discordgo.Session
	conf *Config
}

func NewCommandParser(dg *discordgo.Session) (parser *CommandParser){
	parser = &CommandParser{dg: dg}
	return parser
}

func (h *CommandParser) Read(s *discordgo.Session, m *discordgo.MessageCreate) {
	// very important to set this first!
	cp := h.conf.BotConfig.CP

	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore bots
	if m.Author.Bot {
		return
	}

	message := strings.Fields(strings.ToLower(m.Content))
	if len(message) < 1 {
		log.Println(m.Content)
		return
	}

	command := strings.ToLower(message[0])
	// If the message is "ping" reply with "Pong!"
	if command == cp+"ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
		return
	}


}