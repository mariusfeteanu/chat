package main

import (
	"encoding/json"
	"io"
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
			fw := assertCanFlush(w)
			uch := ensureUserChannel(&channels, user)
			ha := make(chan byte)
			defer func() {
				close(ha)
			}()

			go uch.heartbeat(ha, 1000)
			uch.receive(fw)
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

type flushWriter interface {
	io.Writer
	http.Flusher
}

func assertCanFlush(w io.Writer) (f flushWriter) {
	_, cantFlushErr := w.(http.Flusher)
	if !cantFlushErr {
		panic(cantFlushErr)
	}
	return w.(flushWriter)
}

func (uch userChannel) receive(fw flushWriter) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("disconnected (receive panic) [%v]\n", uch.User)
		}
	}()

	fw.Flush()

	log.Printf("receving messages for [%v]\n", uch.User)
	for {
		msg := <-uch.Channel
		if msg.From == "" {
			_, err := fw.Write([]byte{heartbyte})
			if err != nil {
				log.Printf("disconnected (heartbeat) [%v]\n", uch.User)
				return
			}
			fw.Flush()
			continue
		}

		jsonBytes, _ := json.Marshal(msg)
		jsonBytes = append(jsonBytes, heartbyte)
		_, err := fw.Write(jsonBytes)
		if err == nil {
			log.Printf("receving for [%v]\n", uch.User)
			fw.Flush()
		} else {
			log.Printf("disconnected (receive) [%v]\n", uch.User)
			return
		}
	}
}

func (uch userChannel) send(from string, msg []byte) {
	log.Printf("sending from [%v]\n", from)
	uch.Channel <- message{from, string(msg)}
}

func (uch userChannel) heartbeat(ha chan byte, ms uint) {
	for {
		uch.Channel <- message{From: "", Content: ""}
		select {
		case <-ha:
			log.Println("hearbeat stopped for: " + uch.User)
			return
		default:
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}
}
