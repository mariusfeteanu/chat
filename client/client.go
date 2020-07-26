package main

import (
    "net/http"
    "fmt"
    "log"
    "strings"
    "bufio"
    "os"
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

        post_url := "http://localhost:8080/messages"

        req, _ := http.NewRequest("POST", post_url, strings.NewReader(message))
        req.Header.Add("Chat-From", username)
        req.Header.Add("Chat-To", to)

        _, err := client.Do(req)
        if err != nil {
            log.Println(err)
        }
        fmt.Println(post_url)
    }
}
