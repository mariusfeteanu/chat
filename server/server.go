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

func getIncomingChannel(userChannels *sync.Map, u string) (userChannel chan userMessage) {
	hopeUserChannel, exists := userChannels.Load(u)
	if !exists {
		userChannel = make(chan userMessage, 100)
		userChannels.Store(u, userChannel)
	} else {
		userChannel = hopeUserChannel.(chan userMessage)
	}
	return userChannel
}

func dispatchUserIncomingMessages(userChannels *sync.Map, u string, w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("disconnected (receive goroutine): " + u)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	for {
		userChannel := getIncomingChannel(userChannels, u)
		msg := <-userChannel
		jsonBytes, _ := json.Marshal(msg)
		_, err := w.Write(jsonBytes)
		if err == nil {
			f.Flush()
		} else {
			log.Println("disconnected : " + u)
			return
		}
	}
}

func sendMessage(userChannels *sync.Map, from string, to string, message []byte) {
	userChannel := getIncomingChannel(userChannels, to)
	userChannel <- userMessage{from, string(message)}
}

func keepAlive(u string, w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("disconnected (receive goroutine): " + u)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	for {
		jsonBytes, jsonError := json.Marshal(nil)
		if jsonError != nil {
			log.Println(jsonError)
		}
		_, err := w.Write(jsonBytes)
		if err == nil {
			f.Flush()
		} else {
			log.Println("disconnected (heartbeat) : " + u)
			return
		}
		time.Sleep(time.Second)
	}
}

func main() {
	log.Println("Starting server")

	var userChannels sync.Map

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		user := pathParts[len(pathParts)-1]

		if r.Method == "POST" {
			to := r.Header.Get("Chat-To")
			message, _ := ioutil.ReadAll(r.Body)
			sendMessage(&userChannels, user, to, message)
		}

		if r.Method == "GET" {
			log.Println("getting messages for " + user)
			go dispatchUserIncomingMessages(&userChannels, user, w)
			keepAlive(user, w)
		}
	})

	certFile := "server.crt"
	keyFile := "server.key"
	log.Fatal(http.ListenAndServeTLS(":8080", certFile, keyFile, nil))
}
