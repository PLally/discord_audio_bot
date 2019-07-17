package main

import (
	"github.com/bwmarrin/discordgo"
	"sync"
)

type VoiceManager struct {
	sync.RWMutex
	connections map[*discordgo.VoiceConnection]*OpenVoiceConnection
}

func (m *VoiceManager) Set(key *discordgo.VoiceConnection, value *OpenVoiceConnection) {
	m.Lock()
	m.connections[key] = value
	m.Unlock()
}


func (m *VoiceManager) Get(key *discordgo.VoiceConnection) (value *OpenVoiceConnection, ok bool) {
	m.RLock()
	result, ok := m.connections[key]
	m.RUnlock()
	return result, ok
}



var vcManager = VoiceManager{
	connections: make(map[*discordgo.VoiceConnection]*OpenVoiceConnection),
}

func newOpenVoiceConnection(vc *discordgo.VoiceConnection)  (*OpenVoiceConnection) {
	ovc := &OpenVoiceConnection{
		VoiceConnection: vc,
		recv: make(chan *discordgo.Packet, 300),
		voiceUsers: make(map[string]*VoiceUser),
		userLookup: make(map[uint32]*VoiceUser),
	}
	return ovc
}
type OpenVoiceConnection struct {
	VoiceConnection *discordgo.VoiceConnection
	recv chan *discordgo.Packet
	voiceUsers map[string]*VoiceUser
	userLookup map[uint32]*VoiceUser
	userMapLock sync.RWMutex
}

type VoiceUser struct {
	SSRC uint32
	UserID string

}

func (ovc *OpenVoiceConnection) Close() {
	ovc.VoiceConnection.Close()
}

func (ovc *OpenVoiceConnection) newUser(userID string) (user *VoiceUser) {
	user = &VoiceUser {
		SSRC: 0,
		UserID: userID,
	}

	ovc.userMapLock.Lock()
	ovc.voiceUsers[userID] = user
	ovc.userMapLock.Unlock()

	return user
}

func (ovc *OpenVoiceConnection) setSSRC(userID string, ssrc uint32) {
	ovc.userMapLock.RLock()
	user, ok := ovc.voiceUsers[userID]
	ovc.userMapLock.RUnlock()
	if !ok {
		user = ovc.newUser(userID)
	}
	user.SSRC = ssrc

	ovc.userMapLock.Lock()
	ovc.userLookup[ssrc] = user
	ovc.userMapLock.Unlock()
}

func (ovc *OpenVoiceConnection) getUserFromSSRC(ssrc uint32) (user *VoiceUser, ok bool) {
	ovc.userMapLock.RLock()
	user, ok = ovc.userLookup[ssrc]
	ovc.userMapLock.RUnlock()
	return user, ok
}
