package main

import (
	"github.com/bwmarrin/discordgo"
	"time"
	"github.com/bwmarrin/dgvoice"
	"os"
	"encoding/binary"
	"fmt"
)


func voiceUpdate(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {

	ovc, ok := vcManager.Get(vc)
	if !ok {
		return
	}

	ovc.setSSRC(vs.UserID, uint32(vs.SSRC))
}

func listen(vc *discordgo.VoiceConnection, users []string) ( *OpenVoiceConnection ){
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

	ovc := newOpenVoiceConnection(vc)

	vcManager.Set(vc, ovc)

	vc.AddHandler(voiceUpdate)

	go dgvoice.ReceivePCM(vc, ovc.recv)
	go func() {
		files := make(map[string]*os.File)
		for _, id := range users {

			filepath := fmt.Sprintf("recorded_audio/users/%s/",id)
			_ = os.MkdirAll(filepath,  os.ModePerm)
			f, err := os.Create(filepath+ time.Now().Format("2006-01-02 15-04-05")+".pcm")

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
			user, ok := ovc.getUserFromSSRC(data.SSRC)
			if !ok {
				continue
			}
			if isListenedTo[user.UserID] {
				bytes := make([]byte, len(data.PCM)*2)
				for i, n := range data.PCM {
					binary.LittleEndian.PutUint16(bytes[i*2:], uint16(n))
				}
				fmt.Println(len(data.PCM))
				f := files[user.UserID]
				if f != nil {
					f.Write(bytes)
				}

			}
		}

	}()
	return nil
}
