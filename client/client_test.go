package main

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestReadLine(t *testing.T) {
	cases := []struct {
		command,
		readCommand string
	}{
		{"what\n", "what"},
		{"what\n\r", "what"},
	}

	for _, c := range cases {
		comReader := strings.NewReader(c.command)
		comBuffered := bufio.NewReader(comReader)

		readCommand := readLine(comBuffered)

		if readCommand != "what" {
			t.Errorf("%v != %v", c.readCommand, c.command)
		}
	}
}

type ClientMock struct {
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	body := ioutil.NopCloser(strings.NewReader("{\"From\": \"u1\", \"Content\": \"ohai\"}"))
	return &http.Response{Body: body}, nil
}

func TestReceive(t *testing.T) {
	urlExample := "https://example.com/messages/u1"
	fromExample := "u1"
	contentExample := "ohai"

	checkShow := func(from string, content string) {
		if from != fromExample {
			t.Errorf("%v != %v", from, fromExample)
		}
		if content != contentExample {
			t.Errorf("%v != %v", content, contentExample)
		}
	}

	receive(urlExample, &ClientMock{}, checkShow)
}
