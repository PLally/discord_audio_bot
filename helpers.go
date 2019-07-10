package main

import (
	"github.com/bwmarrin/discordgo"
)

func getVoiceState(guild *discordgo.Guild, user *discordgo.User) *discordgo.VoiceState {
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == user.ID {
			return voiceState
		}
	}
	return nil
}

func sendSilence(vc *discordgo.VoiceConnection, amount int) {
	for i := 1; i <= amount; i++ {
		vc.OpusSend <- []byte{0xF8, 0xFF, 0xFE}
	}
}
