package main

import (
	"github.com/bwmarrin/discordgo"
	"time"
	"github.com/bwmarrin/dgvoice"
	"os"
	"encoding/binary"
	"fmt"
)
var openVoiceConnections = make(map[*discordgo.VoiceConnection]openVoiceConnection)
func voiceUpdate(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {
	ovc := openVoiceConnections[vc]
	ssrc := uint32(vs.SSRC)
	ovc.users[vs.UserID] = ssrc
	ovc.userLookup[ssrc] = vs.UserID
}

func listenToFiles(vc *discordgo.VoiceConnection, users []string) ( *openVoiceConnection ){
	isListenedTo := make(map[string]bool)
	for _, u := range users {
		isListenedTo[u] = true
	}

	go func() {
		vc.Speaking(true)
		sendSilence(vc, 4)
		time.Sleep(time.Millisecond * 100)
		vc.Speaking(false)
	}()

	ovc := openVoiceConnection{
		voiceConnection: vc,
		recv: make(chan *discordgo.Packet, 300),
		users: make(map[string]uint32),
		userLookup: make(map[uint32]string),
	}

	openVoiceConnections[vc] = ovc
	vc.AddHandler(voiceUpdate)

	go dgvoice.ReceivePCM(vc, ovc.recv)
	go func() {
		files := make(map[string]*os.File)
		for _, id := range users {

			filepath := fmt.Sprintf("recorded_audio/users/%s/",id)
			_ = os.MkdirAll(filepath,  os.ModePerm)
			f, err := os.Create(filepath+ time.Now().Format("2006-01-02 15-04-05")+".pcm")

			fmt.Println(err)
			if err == nil {
				files[id] = f
			} else {
				isListenedTo[id] = false
			}
		}


		defer func() {
			for _, f := range files {
				f.Close()
			}
		}()

		for {
			data, ok := <-ovc.recv
			if !ok {
				return
			}
			userID := ovc.userLookup[data.SSRC]
			if isListenedTo[userID] {
				bytes := make([]byte, len(data.PCM)*2)
				for i, n := range data.PCM {
					binary.LittleEndian.PutUint16(bytes[i*2:], uint16(n))
				}

				f := files[userID]
				if f != nil {
					f.Write(bytes)
				}

			}
		}

	}()
	return nil
}

type openVoiceConnection struct {
	voiceConnection *discordgo.VoiceConnection
	recv chan *discordgo.Packet
	users map[string]uint32
	userLookup map[uint32]string
}
