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

var heartbyte = byte(0)

func main() {
	log.Println("Starting server")

	var channels sync.Map

	http.HandleFunc("/messages/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v: %v", r.Method, r.URL.Path)
		pathParts := strings.Split(r.URL.Path, "/")
		user := pathParts[len(pathParts)-1]

		if r.Method == "POST" {
			to := r.Header.Get("Chat-To")
			message, _ := ioutil.ReadAll(r.Body)
			uch := ensureUserChannel(&channels, to)
			uch.send(user, message)
		}

		if r.Method == "GET" {
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
		channels.Store(user, ch)
		log.Printf("established channel for [%v]\n", user)
	} else {
		ch = someChannel.(chan message)
	}
	return userChannel{user, ch}
}

func (uch userChannel) receive(w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("disconnected (receive panic) [%v]\n", uch.User)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	log.Printf("receving messages for [%v]\n", uch.User)
	for {
		msg := <-uch.Channel
		jsonBytes, _ := json.Marshal(msg)
		jsonBytes = append(jsonBytes, heartbyte)
		_, err := w.Write(jsonBytes)
		if err == nil {
			log.Printf("receving for [%v]\n", uch.User)
			f.Flush()
		} else {
			log.Printf("disconnected (receive) [%v]\n", uch.User)
			return
		}
	}
}

func (uch userChannel) send(from string, msg []byte) {
	log.Printf("sending [%v] -> [%v]\n", from, uch.User)
	uch.Channel <- message{from, string(msg)}
}

func (uch userChannel) keepAlive(w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("disconnected (keep alive panic) [%v]\n", uch.User)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	for {
		_, err := w.Write([]byte{heartbyte})
		if err == nil {
			f.Flush()
		} else {
			log.Printf("disconnected (heartbyte) [%v]\n", uch.User)
			return
		}
		time.Sleep(time.Second)
	}
}
