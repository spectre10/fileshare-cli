package receive

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pion/webrtc/v3"
	"github.com/spectre10/fileshare-cli/lib"
)

func (s *Session) HandleState() {
	s.peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		fmt.Printf("\nICE Connection State has changed: %s\n\n", state.String())
		if state == webrtc.ICEConnectionStateDisconnected {
			s.done <- struct{}{}
		}
	})
	s.peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		if dc.Label() == "control" {
			s.controlChannel = dc
			s.assign(dc)
		} else {
			s.transferChannel = dc
			s.transferChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
				s.msgChan <- msg.Data
			})
		}
	})
}

func (s *Session) assign(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		// fmt.Printf("New Data Channel Opened! '%s' - '%d'\n", dc.Label(), dc.ID())
	})
	dc.OnClose(func() {
		fmt.Println("Channel Closed!")
		s.close(true)
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		if !s.sizeDone {
			var md lib.Metadata
			err := json.Unmarshal(msg.Data, &md)
			if err != nil {
				panic(err)
			}
			s.size = md.Size
			s.name = md.Name
			s.file, err = os.OpenFile(s.name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				panic(err)
			}
			s.sizeDone = true
			var consent string
			fmt.Printf("Do you want to receive '%s' ? [Y/N] ", s.name)
			fmt.Scanln(&consent)
			if consent == "n" || consent == "N" {
				s.controlChannel.SendText("n")
			} else {
				s.controlChannel.SendText("Y")
				s.consentChan <- struct{}{}
			}
		}
		//       else {
		// 	s.msgChan <- msg.Data
		// }

	})
}

func (s *Session) close(isOnClose bool) {
	close(s.consentChan)
	close(s.done)
}
