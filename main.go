package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)
var TOKEN = os.Getenv("DISCORD_BOT_TOKEN")

func main() {
	dg, err := discordgo.New("Bot "+TOKEN)
	if err != nil {
		fmt.Println(err)
		return
	}

	dg.AddHandler(messageCreate)

	addCommandFunc("record", recordCommand)
	addCommandFunc("ping", pingCommand)
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
func pingCommand(cmd textCommand, s *discordgo.Session, m *discordgo.MessageCreate) (reply string) {
	timeSent, err := time.Parse( time.RFC3339Nano, string(m.Timestamp))
	if err != nil {
		fmt.Println(err)
		return "pong..."
	}
	delay := ( time.Now().UnixNano() - timeSent.UnixNano()) / 1e6

	return fmt.Sprintf("pong... %v ms", delay)
}

func recordCommand(cmd textCommand, s *discordgo.Session, m *discordgo.MessageCreate) (reply string) {
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return "couldn't find channel"
	}
	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return "couldn't find your guild"
	}

	voiceState := getVoiceState(guild, m.Author)
	if voiceState == nil {
		return "Make sure you're in a voice channel"
	}

	vc, err := s.ChannelVoiceJoin(guild.ID, voiceState.ChannelID, false, false)

	if err != nil {
		return "Couldn't join your voice channel"
	}
	_, ok := vcManager.Get(vc)
	if ok {
		return "The bot is already recording in this guild"
	}
	s.ChannelMessageSend(m.ChannelID, "Joined your voice Channel")

	var users []string
	for _, m := range m.Mentions {
		users = append(users, m.ID)
	}
	listen(vc, users)
	return "Recording users"
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}
	checkCommands(s, m)

}
