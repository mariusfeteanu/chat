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
	Content string
}

type userChannel struct {
	User    string
	Channel chan message
}

func main() {
	log.Println("Starting server")

	var channels sync.Map

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		user := pathParts[len(pathParts)-1]

		if r.Method == "POST" {
			to := r.Header.Get("Chat-To")
			message, _ := ioutil.ReadAll(r.Body)
			uch := ensureUserChannel(&channels, to)
			uch.send(user, message)
		}

		if r.Method == "GET" {
			log.Println("getting messages for " + user)
			uch := ensureUserChannel(&channels, user)
			go uch.receive(w)
			uch.keepAlive(w)
		}
	})

	certFile := "server.crt"
	keyFile := "server.key"
	log.Fatal(http.ListenAndServeTLS(":8080", certFile, keyFile, nil))
}

func ensureUserChannel(channels *sync.Map, user string) (uch userChannel) {
	someChannel, exists := channels.Load(user)
	var ch chan message
	if !exists {
		ch = make(chan message, 100)
		channels.Store(uch.User, ch)
	} else {
		ch = someChannel.(chan message)
	}
	return userChannel{user, ch}
}

func (uch userChannel) receive(w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("disconnected (receive panic): " + uch.User)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	for {
		msg := <-uch.Channel
		jsonBytes, _ := json.Marshal(msg)
		_, err := w.Write(jsonBytes)
		if err == nil {
			f.Flush()
		} else {
			log.Println("disconnected (receive): " + uch.User)
			return
		}
	}
}

func (uch userChannel) send(from string, msg []byte) {
	uch.Channel <- message{from, string(msg)}
}

func (uch userChannel) keepAlive(w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("disconnected (keep alive panic): " + uch.User)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	for {
		js, err := json.Marshal(nil)
		if err != nil {
			log.Fatalf("disconnected (heartbeat error): %v", err)
			panic(err)
		}
		_, err = w.Write(js)
		if err == nil {
			f.Flush()
		} else {
			log.Println("disconnected (heartbeat) : " + uch.User)
			return
		}
		time.Sleep(time.Second)
	}
}
