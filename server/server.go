package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type message struct {
	From    string
	To      string
	Content string
}

type userMessage struct {
	From    string
	Content string
}

func main() {
	log.Println("Starting server")

	var userChannels sync.Map

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		username := pathParts[len(pathParts)-1]

		if r.Method == "POST" {
			to := r.Header.Get("Chat-To")
			log.Println("[" + username + "] -> [" + to + "]")
			messageBytes, _ := ioutil.ReadAll(r.Body)
			msg := message{
				username,
				to,
				string(messageBytes)}

			var userChannel chan userMessage
			hopeUserChannel, exists := userChannels.Load(msg.To)
			if !exists {
				userChannel = make(chan userMessage, 100)
				userChannels.Store(msg.To, userChannel)
			} else {
				userChannel = hopeUserChannel.(chan userMessage)
			}
			userChannel <- userMessage{msg.From, msg.Content}
		}

		if r.Method == "GET" {
			log.Println("getting messages for " + username)
			flusher, _ := w.(http.Flusher)
			flusher.Flush()

			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("disconnected (receive goroutine): " + username)
					}
				}()
				var userChannel chan userMessage
				hopeUserChannel, exists := userChannels.Load(username)
				if !exists {
					userChannel = make(chan userMessage, 100)
					userChannels.Store(username, userChannel)
				} else {
					userChannel = hopeUserChannel.(chan userMessage)
				}
				for {
					msg := <-userChannel
					jsonBytes, _ := json.Marshal(msg)
					_, err := w.Write(jsonBytes)
					if err == nil {
						flusher.Flush()
					} else {
						log.Println("disconnected : " + username)
						return
					}
				}
			}()

			for {
				jsonBytes, jsonError := json.Marshal(nil)
				if jsonError != nil {
					log.Println(jsonError)
				}
				_, err := w.Write(jsonBytes)
				if err == nil {
					flusher.Flush()
				} else {
					log.Println("disconnected (heartbeat) : " + username)
					return
				}
				time.Sleep(time.Second)
			}
		}
	})

	certFile := "server.crt"
	keyFile := "server.key"
	log.Fatal(http.ListenAndServeTLS(":8080", certFile, keyFile, nil))
}
