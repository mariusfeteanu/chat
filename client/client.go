package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var heartbyte = byte(0)

func main() {
	in := bufio.NewReader(os.Stdin)

	// user info
	var username string
	prompt()
	username = readLine(in, "username: ")
	url := "https://localhost:8080/messages/" + username

	// http client
	client := createClient()

	// start receiving messages
	go receive(url, client, showCliMessage)

	// user interaction loop
	var to string
	for {
		var message string
		message = readLine(in)

		if message == "/q" {
			break
		}

		if strings.HasPrefix(message, "/t") {
			to = message[3:]
			info(fmt.Sprintf("sending to: <<%v>>", to))
			continue
		}

		if to == "" {
			info("try /t some_user_name, to talk to someone")
			continue
		}

		prompt()
		req, _ := http.NewRequest("POST", url, strings.NewReader(message))
		req.Header.Add("Chat-To", to)

		_, err := client.Do(req)
		if err != nil {
			info(fmt.Sprintf("ERROR: %v", err))
		}
	}
}

func readLine(r *bufio.Reader, prompts ...string) string {
	for _, p := range prompts {
		fmt.Print(p)
	}
	line, err := r.ReadString('\n')
	if err != nil {
		panic(err)
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
}

func info(msg string) {
	fmt.Println("# " + msg)
	prompt()
}

func prompt() {
	fmt.Print("~ ")
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func receive(url string, client httpClient, showFunc func(string, string)) {
	pr, _ := io.Pipe()
	req, err := http.NewRequest("GET", url, ioutil.NopCloser(pr))
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	buff := make([]byte, 1024)
	for {
		n, rerr := resp.Body.Read(buff)
		if rerr != nil {
			if rerr == io.EOF {
				info("server disconnected (EOF)")
				return
			}
			panic(rerr)
		}
		raw := buff[:n]

		rawMessages := bytes.Split(raw, []byte{heartbyte})

		for _, rawm := range rawMessages {

			if len(rawm) == 0 { // heartbeat
				continue
			}

			if n > 0 {
				var message map[string]interface{}
				err := json.Unmarshal(rawm, &message)
				if err != nil {
					panic(err)
				}
				if message != nil {
					showFunc(message["From"].(string), message["Content"].(string))
				}
			}
		}
	}
}

func showCliMessage(from string, content string) {
	fmt.Printf(
		"[%v]: %v\n",
		from,
		content)
	prompt()
}

func createClient() *http.Client {
	info("starting client")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, ForceAttemptHTTP2: true,
	}
	client := &http.Client{Transport: tr}
	return client
}
