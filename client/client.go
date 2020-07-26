package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"io/ioutil"
	"time"
)

func read_line(r *bufio.Reader) string {
	line, err := r.ReadString('\n')
	if err == nil {
		return strings.TrimSuffix(line, "\n")
	} else {
		return ""
	}
}

func main() {
	in := bufio.NewReader(os.Stdin)
	log.Println("Starting client")
	client := &http.Client{}

	fmt.Print("username: ")
	var username string
	username = read_line(in)
	fmt.Printf("<<%v>>\n", username)
    post_url := "http://localhost:8080/messages/" + username

	go func() {
        for {
            resp, err := client.Get(post_url)
            if resp != nil {
                bytes_message, _ := ioutil.ReadAll(resp.Body)
                if len(bytes_message) > 0 {
                    text_message := string(bytes_message)
                    fmt.Println(text_message)
            		fmt.Print("~ ")
                }
            } else {
                fmt.Println("ERROR:", err)
        		fmt.Print("~ ")
            }
            time.Sleep(time.Second)
        }
	}()

	var to string

	for {
		var message string
		fmt.Print("~ ")

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
			fmt.Println("Try /t some_user_name, to talk to someone")
			continue
		}

		req, _ := http.NewRequest("POST", post_url, strings.NewReader(message))
		req.Header.Add("Chat-From", username)  // this is reduntant
		req.Header.Add("Chat-To", to)

		_, err := client.Do(req)
		if err != nil {
			log.Println("ERROR: ", err)
		}
	}
}
