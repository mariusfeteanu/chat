package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	// user info
	var username string
	read_line := func(r *bufio.Reader, prompt ...string) string {
		fmt.Printf("%v |> ", username)
		for _, p := range prompt {
			fmt.Print(p)
		}
		line, err := r.ReadString('\n')
		if err == nil {
			return strings.TrimSuffix(line, "\n")
		} else {
			return ""
		}
	}
	username = read_line(in, "username: ")
	url := "https://localhost:8080/messages/" + username

	// client setup
	fmt.Println("# Starting client")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, ForceAttemptHTTP2: true,
	}
	client := &http.Client{Transport: tr}

	// receive messages
	pr, _ := io.Pipe()
	req, _ := http.NewRequest("GET", url, ioutil.NopCloser(pr))
	resp, _ := client.Do(req)
	go func() {
		for {
			b := make([]byte, 1024)
			n, _ := resp.Body.Read(b)
			if n > 0 {
				fmt.Printf("\n%v%v |> ", string(b), username)
			}
		}
	}()

	// user interaction loop
	var to string
	for {
		var message string

		message = read_line(in)
		if message == "/q" {
			break
		}

		if strings.HasPrefix(message, "/t") {
			to = message[3:]
			fmt.Println(fmt.Sprintf("# sending to: <<%v>>", to))
			continue
		}

		if to == "" {
			fmt.Println("# Try /t some_user_name, to talk to someone")
			continue
		}

		req, _ := http.NewRequest("POST", url, strings.NewReader(message))
		req.Header.Add("Chat-To", to)

		_, err := client.Do(req)
		if err != nil {
			fmt.Println("# ERROR: ", err)
		}
	}
}
