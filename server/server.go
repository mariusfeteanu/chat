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

type userContext struct {
	User        string
	allChannels *sync.Map
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
			userContext{to, &channels}.send(user, message)
		}

		if r.Method == "GET" {
			log.Println("getting messages for " + user)
			ctx := userContext{user, &channels}
			go ctx.receive(w)
			ctx.keepAlive(w)
		}
	})

	certFile := "server.crt"
	keyFile := "server.key"
	log.Fatal(http.ListenAndServeTLS(":8080", certFile, keyFile, nil))
}

func (uch userContext) incomingCh() (ch chan message) {
	someChannel, exists := uch.allChannels.Load(uch.User)
	if !exists {
		ch = make(chan message, 100)
		uch.allChannels.Store(uch.User, ch)
	} else {
		ch = someChannel.(chan message)
	}
	return ch
}

func (uch userContext) receive(w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("disconnected (receive goroutine): " + uch.User)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	for {
		ch := uch.incomingCh()
		msg := <-ch
		jsonBytes, _ := json.Marshal(msg)
		_, err := w.Write(jsonBytes)
		if err == nil {
			f.Flush()
		} else {
			log.Println("disconnected : " + uch.User)
			return
		}
	}
}

func (uch userContext) send(from string, msg []byte) {
	ch := uch.incomingCh()
	ch <- message{from, string(msg)}
}

func (uch userContext) keepAlive(w http.ResponseWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("disconnected (receive goroutine): " + uch.User)
		}
	}()

	f, _ := w.(http.Flusher)
	f.Flush()

	for {
		js, err := json.Marshal(nil)
		if err != nil {
			log.Fatalf("disconnected (error): %v", err)
			return
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
