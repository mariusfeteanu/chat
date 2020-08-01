package main

import (
	"errors"
	"testing"
	"time"
)

func TestHearbeat(t *testing.T) {
	quit := make(chan byte)      // channel to kill the heartbeat
	ch := make(chan message, 10) // channel to receive heartbeat messages
	uch := userChannel{Channel: ch, User: "someone"}

	go uch.heartbeat(quit, 1)
	time.Sleep(time.Duration(3) * time.Millisecond)
	quit <- byte(0)

	select {
	case <-ch:
		t.Log("ok")
	default:
		t.Error("no heartbeat")
	}
}

func TestSend(t *testing.T) {
	ch := make(chan message, 10)
	uch := userChannel{Channel: ch, User: "to"}

	uch.send("from", []byte("hello"))

	select {
	case msg := <-ch:
		if msg.Content != "hello" {
			t.Errorf("%s != hello", msg.Content)
		}
	default:
		t.Error("no message sent")
	}

}

type flushWriterMock struct {
	buffer  *[]byte
	flushes *uint
}

func (fw flushWriterMock) Write(p []byte) (n int, err error) {
	if len(p) == 1 && p[0] == heartbyte {
		return 0, errors.New("done")
	}

	*fw.buffer = append(*fw.buffer, p...)

	return len(p), nil
}

func (fw flushWriterMock) Flush() {
	*fw.flushes++
}

func TestReceive(t *testing.T) {
	ch := make(chan message, 10)
	uch := userChannel{Channel: ch, User: "to"}

	uch.send("from", []byte("hello"))
	uch.send("from", []byte("there"))
	uch.send("", []byte{}) // same as heartbeat end

	b := []byte{}
	f := new(uint)
	wf := flushWriterMock{&b, f}
	uch.receive(wf)

	expBuffer := "{\"From\":\"from\",\"Content\":\"hello\"}\x00{\"From\":\"from\",\"Content\":\"there\"}\x00"
	buffer := string(*wf.buffer)
	if buffer != expBuffer {
		t.Errorf("'%s' != '%s'", buffer, expBuffer)
	}

	if *wf.flushes != 3 {
		t.Errorf("Expected exactly one flush per message. %v != %v", *wf.flushes, 3)
	}
}
