package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

type message struct {
	from    string
	to      string
	content string
}

type user_message struct {
	from    string
	content string
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
			msg := message{
				username,
				to,
				string(message_bytes)}

			var user_channel chan user_message
			hope_user_channel, exists := user_channels.Load(msg.to)
			if !exists {
				user_channel = make(chan user_message, 100)
				user_channels.Store(msg.to, user_channel)
			} else {
				user_channel = hope_user_channel.(chan user_message)
			}
			user_channel <- user_message{msg.from, msg.content}
		}

		if r.Method == "GET" {
			log.Println("getting messages for " + username)
			flusher, _ := w.(http.Flusher)
			flusher.Flush()
			for {
				hope_user_channel, exists := user_channels.Load(username)
				if !exists { // it's fine, no messages yet
					continue
				}
				user_channel := hope_user_channel.(chan user_message)
				select {
				case msg := <-user_channel:
					log.Println("[" + username + "] <- [" + msg.from + "]")
					text_message := msg.from + " |< " + msg.content + "\n"
					w.Write([]byte(text_message))
					flusher.Flush()
				default:
					continue
				}
			}
		}
	})

	cert_file := "server.crt"
	key_file := "server.key"
	log.Fatal(http.ListenAndServeTLS(":8080", cert_file, key_file, nil))
}
