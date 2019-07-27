package main

import (
	"encoding/binary"
	"fmt"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"io"
	"io/ioutil"
	"os"
	"time"
)

func voiceUpdate(vc *discordgo.VoiceConnection, vs *discordgo.VoiceSpeakingUpdate) {

	ovc, ok := vcManager.Get(vc)
	if !ok {
		return
	}
	user, ok := ovc.getUser(vs.UserID)
	if !ok {
		path := fmt.Sprintf("recorded_audio/users/%s/", vs.UserID)
		var f io.Writer
		_ = os.MkdirAll(path, os.ModePerm)
		f, err := os.Create(path + time.Now().Format("2006-01-02 15-04-05") + ".pcm")
		if err != nil {
			f = ioutil.Discard
		}

		user = ovc.newUser(vs.UserID, f)
	}
	ovc.setSSRC(user, uint32(vs.SSRC))
}

func listen(vc *discordgo.VoiceConnection, users []string) *OpenVoiceConnection {
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
		for {
			data, ok := <-ovc.recv
			if !ok {
				return
			}
			user, ok := ovc.getUserFromSSRC(data.SSRC)
			if !ok {
				continue
			}
			s := 0 // amount of silence since last packet
			if isListenedTo[user.UserID] {

				if user.lastPacket != nil && ovc.recordSilence {
					s = int((data.Timestamp - user.lastPacket.Timestamp - 960) * 2)
				}
				bytes := make([]byte, len(data.PCM)*2+s)
				for i, n := range data.PCM {
					binary.LittleEndian.PutUint16(bytes[i*2+s:], uint16(n))
				}
				user.audio.Write(bytes)
			}
			user.lastPacket = data

		}

	}()
	return ovc
}
