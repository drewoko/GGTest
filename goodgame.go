package main

import (
	"sync"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"encoding/json"
	"fmt"
)

type GoodGameSocketStorage struct {
	sync.Mutex
	wsClient *websocket.Conn
}

func (s *GoodGameSocketStorage) writeMessage(request []byte) error {
	s.Lock()
	defer s.Unlock()

	return s.wsClient.WriteMessage(websocket.TextMessage, request)
}

func initGoodGame(wg *sync.WaitGroup) {
	wg.Add(1)
	log.Println("Connecting to GoodGame.ru")

	wsClient, _, err := websocket.DefaultDialer.Dial("wss://chat.goodgame.ru/chat/websocket", nil)

	if err != nil {
		log.Println("Failed to connect to GoodGame.ru", err)
		time.Sleep(time.Second * 5)
		initGoodGame(wg)
		return
	}

	socket := &GoodGameSocketStorage{
		wsClient: wsClient,
	}

	plainMessageChan := make(chan []byte)
	channelChan := make(chan string)
	quitChat := make(chan bool)

	pingTicker := time.NewTicker(time.Second * 1)
	viewingTicker := time.NewTicker(time.Second * 5)
	defer func() {
		close(plainMessageChan)
		close(channelChan)
		pingTicker.Stop()
		viewingTicker.Stop()
		wg.Done()
	}()

	go func() {
		for {
			messageType, message, err := wsClient.ReadMessage()

			if err != nil {
				log.Println("Disconnected from GoodGame.ru")
				quitChat <- true
				time.Sleep(time.Second * 5)
				initGoodGame(wg)
				return
			}

			if messageType == websocket.TextMessage {
				plainMessageChan <- message
			}
		}
	}()

	for {
		var plainMessage []byte
		select {
		case plainMessage = <-plainMessageChan:

			message := GoodGameStruct{}

			json.Unmarshal(plainMessage, &message)

			fmt.Println(message)

		case <-pingTicker.C:
			sendPing(socket)
		case <-pingTicker.C:
			sendViewing(socket)
		case <-quitChat:
			return
		}
	}
}



//"type":"viewing","data":{"channel":"3893","userId":"","newplayer":1}}
func sendViewing(socket *GoodGameSocketStorage) {

	data := make(map[string]interface{})

	data["channel"] = CHAN
	data["userId"] = ""
	data["newplayer"] = 1

	sentMessage(socket, GoodGameStruct{
		Type: "viewing",
		Data: data,
	})
}


func sendPing(socket *GoodGameSocketStorage) {

	data := make(map[string]interface{})

	data["channel"] = CHAN
	data["state"] = 1

	sentMessage(socket, GoodGameStruct{
		Type: "ping",
		Data: data,
	})
}

func joinToChannel(socket *GoodGameSocketStorage, channel interface{}) {
	sentMessage(socket, GoodGameStruct{
		Type: "join",
		Data: map[string]interface{}{"channel_id": channel, "hidden": false},
	})
}

func requestChannels(socket *GoodGameSocketStorage, start int, count int) {
	sentMessage(socket, GoodGameStruct{
		Type: "get_channels_list",
		Data: map[string]interface{}{"start": start, "count": count},
	})
}

func sentMessage(socket *GoodGameSocketStorage, messageStruct GoodGameStruct) {

	request, err := json.Marshal(messageStruct)

	if err != nil {
		log.Println("Failed to create JSON", err)
		return
	}

	err = socket.writeMessage(request)

	if err != nil {
		log.Println(err)
	}
}

type GoodGameStruct struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
