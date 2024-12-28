package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"onichan/model"
	"onichan/utils"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
)

type MessagePayload struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn   *websocket.Conn
	PostID uint
}

var Users = make(map[uint]Client)
var Posts = make(map[uint]map[uint]bool)
var mu sync.Mutex

func WsHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	jwtToken, err := utils.ValidateJWT(token)
	if err != nil || !jwtToken.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	claims := jwtToken.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade: %v", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	client := Client{Conn: conn}
	Users[userID] = client
	mu.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		var payload MessagePayload
		if err := json.Unmarshal(message, &payload); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			break
		}

		if payload.Type == "post" {
			postID, _ := strconv.Atoi(payload.Data)
			if Posts[uint(postID)] == nil {
				Posts[uint(postID)] = make(map[uint]bool)
			}
			Posts[uint(postID)][userID] = true
			if Users[userID].PostID != 0 {
				delete(Posts[Users[userID].PostID], userID)
			}
			client := Users[userID]
			client.PostID = uint(postID)
			Users[userID] = client
		}
	}

	mu.Lock()
	delete(Users, userID)
	mu.Unlock()
}

func SendWebSocketNotification(userID uint, message model.Notification) {
	mu.Lock()
	client, ok := Users[userID]
	mu.Unlock()
	if !ok {
		return
	} else {
		if err := client.Conn.WriteJSON(gin.H{
			"data": message,
			"type": "notification",
		}); err != nil {
			log.Printf("Error writing message: %v", err)
		}
	}
}

func SendNewPostSignal(postID uint) {
	mu.Lock()
	clients, ok := Posts[postID]
	mu.Unlock()

	if !ok {
		return
	} else {
		for userID := range clients {
			fmt.Println("user subscribed ", userID)
			if err := Users[userID].Conn.WriteJSON(gin.H{
				"type":    "post",
				"post_id": postID,
			}); err != nil {
				mu.Lock()
				delete(clients, userID)
				mu.Unlock()
				Users[userID].Conn.Close()
			}
		}

		mu.Lock()
		if len(clients) == 0 {
			delete(Posts, postID)
		}
		mu.Unlock()
	}
}
