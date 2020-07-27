package main

import (
	"fmt"
	"net/http"

	//"html"
	//"fmt"
	"io/ioutil"
	"log"
	"strings"
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

	all_incoming := make(chan message, 10)
	user_messages := make(map[string](chan user_message))
	go func() {
		for {
			msg := <-all_incoming
			user_channel, ok := user_messages[msg.to]
			if !ok {
				user_channel = make(chan user_message, 100)
				user_messages[msg.to] = user_channel
			}
			user_channel <- user_message{msg.from, msg.content}
		}
	}()

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			message_bytes, _ := ioutil.ReadAll(r.Body)
			msg := message{
				r.Header.Get("Chat-From"),
				r.Header.Get("Chat-To"),
				string(message_bytes)}
			all_incoming <- msg
		}

		if r.Method == "GET" {
			path_parts := strings.Split(r.URL.Path, "/")
			username := path_parts[len(path_parts)-1]

			select {
			case msg := <-user_messages[username]:
				text_message := "[" + msg.from + "]: " + msg.content
				w.Write([]byte(text_message))
			default:
				if pusher, ok := w.(http.Pusher); ok {
					log.Println("Pushing")
					err := pusher.Push("/pull", nil)
					if err != nil {
						log.Println(fmt.Sprintf("Push Error: %v", err))
					}
				} else {
					log.Println("Can't push")
				}
				w.Write([]byte("nothing"))
				return
			}
		}
	})

	http.HandleFunc("/pull", func(w http.ResponseWriter, r *http.Request) {
		log.Println("pull")
		if r.Method == "GET" {
			log.Println("pull: get")
			w.Write([]byte("my finger"))
		}
	})

	cert_file := "server.crt"
	key_file := "server.key"
	log.Fatal(http.ListenAndServeTLS(":8080", cert_file, key_file, nil))
}
