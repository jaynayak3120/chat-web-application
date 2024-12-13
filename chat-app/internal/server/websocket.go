package server

import (
	model "chat-app/internal/Models"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []model.Message)

// func handleConnections(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		fmt.Fprint(w, err)
// 		return
// 	}
// 	defer conn.Close()
// 	clients[conn] = true
// 	for {
// 		var message string
// 		err := conn.ReadJSON(&message)
// 		if err != nil {
// 			fmt.Fprint(w, err)
// 			delete(clients, conn)
// 			break
// 		}
// 		fmt.Println("Message:", message)
// 		broadcast <- message
// 	}
// }

func HandleMessage() {
	for {
		messages := <-broadcast

		for client := range clients {
			err := client.WriteJSON(&messages)
			if err != nil {
				fmt.Println("Err:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func (s *Server) sendChatHistory(message model.Message) error {
	senderReceiver := make(map[string]string)
	senderReceiver["sender_id"] = message.Sender_Id
	senderReceiver["receiver_id"] = message.Receiver_Id

	messages, err := s.db.GetMessagesforIndividualChat(senderReceiver)
	if err != nil {
		log.Println("Error fetching chat history:", err)
		return err
	}
	broadcast <- messages
	return nil
}
