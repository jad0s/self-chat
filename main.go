package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

type IncomingMessage struct {
	Type        string `json:"type"`
	From        string `json:"from"`
	To          string `json:"to"`
	ContentType string `json:"content_type`
	Content     string `json:"content"`
}

type ActionRequest struct {
	Type     string `json:"type"`
	Action   string `json:"action"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var connections = make(map[string]*websocket.Conn)

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {

		var raw map[string]interface{}
		mt, msg, err := conn.ReadMessage()

		if err != nil {
			log.Println("Error reading message: ", err)
			break
		}

		err = json.Unmarshal(msg, &raw)
		if err != nil {
			log.Println("Invalid JSON: ", err)
			return
		}

		msgType, ok := raw["type"].(string)
		if !ok {
			log.Println("Missing or invalid type")
			return
		}

		switch msgType {
		case "message":
			var parsed IncomingMessage
			err = json.Unmarshal(msg, &parsed)
			if err != nil {
				log.Println("failed to parse message: ", err)
				continue
			}
			fmt.Printf("From %s to %s, %v %v", parsed.From, parsed.To, mt, msg)
			err = conn.WriteMessage(websocket.TextMessage, []byte("Ok"))
			if err != nil {
				log.Println("Error sending message: ", err)
			}

		case "action":
			var parsed ActionRequest
			err := json.Unmarshal(msg, &parsed)
			if err != nil {
				log.Println("Error parsing action request: ", err)
				return
			}
		default:
			log.Println("Unknown message type: ", msgType)

		}

	}
}

func main() {
	scktport := 8081

	dsn := "root:@tcp(localhost:3306)/self-chat"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println("Error connecting to database", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("Error connecting to database", err)
		return
	}

	fmt.Println("db connection successfully established")

	http.HandleFunc("/ws", handler)
	fmt.Println("Running on port", scktport)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", scktport), nil))
}

func register()
