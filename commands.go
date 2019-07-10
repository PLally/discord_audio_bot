package main

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

var commands []textCommand

type commandCallback func(textCommand, *discordgo.Session, *discordgo.MessageCreate) string

type textCommand struct {
	prefix string
	callback commandCallback
}

func addCommand(cmd textCommand) {
	commands = append(commands, cmd)
}

func addCommandFunc(s string, c commandCallback) {
	cmd := textCommand{
		s,
		c,
	}
	addCommand(cmd)
}

func checkCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	for _, cmd := range commands {
		if !strings.HasPrefix(m.Content, cmd.prefix) {
			continue
		}

		reply := cmd.callback(cmd, s, m)
		if reply != "" {
			s.ChannelMessageSend(m.ChannelID, reply)
		}

	}
}
