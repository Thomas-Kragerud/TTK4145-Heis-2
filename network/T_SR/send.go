package T_SR

import (
	"Project/messageHandler"
	"math/rand"
	"time"
)

func Send(
	msgToNetwork chan<- SROnNet,
	msgToSend <-chan messageHandler.NetworkPackage,
) {
	rand.Seed(time.Now().UnixNano())
	var sendingMsg SROnNet
	spam := func(send SROnNet) {
		for i := 0; i < 50; i++ {
			msgToNetwork <- send
			time.Sleep(10 * time.Microsecond)
		}

	}
	for {
		select {
		case msg := <-msgToSend:
			msgId := rand.Intn(90000) + 10000
			sendingMsg = SROnNet{
				Message: msg,
				MsgId:   msgId}
			msgToNetwork <- sendingMsg
			go spam(sendingMsg)
		default:
			time.Sleep(5 * time.Millisecond)
			msgToNetwork <- sendingMsg
		}
	}
}


