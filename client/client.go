package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
			return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
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
	req, err := http.NewRequest("GET", url, ioutil.NopCloser(pr))
	if err != nil {
		log.Fatal(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	go func() {
		for {
			b := make([]byte, 1024)
			n, _ := resp.Body.Read(b)
			if n > 0 {
				var message map[string]interface{}
				err := json.Unmarshal(b[:n], &message)
				if err != nil {
					fmt.Println("ERROR:", err, fmt.Sprintf("<%v>", string(b)))
				}
				if message != nil {
					fmt.Printf(
						"\n%v |< %v\n%v |> ",
						message["From"],
						message["Content"],
						username)
				}
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
