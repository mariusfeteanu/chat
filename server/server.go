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

type Message struct {
	From    string
	To      string
	Content string
}

type UserMessage struct {
	From    string
	Content string
}

func main() {
	log.Println("Starting server")

	var user_channels sync.Map

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		path_parts := strings.Split(r.URL.Path, "/")
		username := path_parts[len(path_parts)-1]

		if r.Method == "POST" {
			to := r.Header.Get("Chat-To")
			log.Println("[" + username + "] -> [" + to + "]")
			message_bytes, _ := ioutil.ReadAll(r.Body)
			msg := Message{
				username,
				to,
				string(message_bytes)}

			var user_channel chan UserMessage
			hope_user_channel, exists := user_channels.Load(msg.To)
			if !exists {
				user_channel = make(chan UserMessage, 100)
				user_channels.Store(msg.To, user_channel)
			} else {
				user_channel = hope_user_channel.(chan UserMessage)
			}
			user_channel <- UserMessage{msg.From, msg.Content}
		}

		if r.Method == "GET" {
			log.Println("getting messages for " + username)
			flusher, _ := w.(http.Flusher)
			flusher.Flush()

			go func() {
				var user_channel chan UserMessage
				hope_user_channel, exists := user_channels.Load(username)
				if !exists {
					user_channel = make(chan UserMessage, 100)
					user_channels.Store(username, user_channel)
				} else {
					user_channel = hope_user_channel.(chan UserMessage)
				}
				for {
					msg := <-user_channel
					json_bytes, _ := json.Marshal(msg)
					_, err := w.Write(json_bytes)
					if err == nil {
						flusher.Flush()
					} else {
						log.Println("disconnected : " + username)
						return
					}
				}
			}()

			for {
				json_bytes, json_error := json.Marshal(nil)
				if json_error != nil {
					log.Println(json_error)
				}
				_, err := w.Write(json_bytes)
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

	cert_file := "server.crt"
	key_file := "server.key"
	log.Fatal(http.ListenAndServeTLS(":8080", cert_file, key_file, nil))
}
