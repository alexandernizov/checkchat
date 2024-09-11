package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexandernizov/checkChat/domain"
)

type Webhook struct {
	log        *slog.Logger
	url        string
	messages   <-chan domain.KafkaMessage
	http       *http.Client
	retryCount int
}

func New(url string, inputChannel <-chan domain.KafkaMessage, retryCount int) *Webhook {
	client := http.Client{}
	return &Webhook{url: url, messages: inputChannel, http: &client, retryCount: retryCount}
}

func (w *Webhook) Serve() {
	go w.ServeChats()
}

func (w *Webhook) ServeChats() {
	for msg := range w.messages {
		jsonData, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("error in marshalling. Msg: %s, %s\n", msg.Key, err.Error())
			continue
		}
		go w.sendChat(msg.Key, jsonData)
	}
}

func (w *Webhook) sendChat(key string, jsonData []byte) {
	for try := 0; try < w.retryCount; try++ {
		req, err := http.NewRequest("POST", "http://"+w.url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("error in connect to http server. Url: %s. Msg: %s, %s. Try: %v\n", w.url, key, err.Error(), try+1)
			try++
			w.log.Warn("chat didn't send because of problems", slog.Attr{Key: "msg.Id", Value: slog.StringValue(key)}, slog.Attr{Key: "Value", Value: slog.StringValue(string(jsonData))})
			time.Sleep(1 * time.Second)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := w.http.Do(req)
		if err != nil {
			fmt.Printf("error in getting response. Msg: %s, %s. Try: %v\n", key, err.Error(), try+1)
			try++
			w.log.Warn("chat didn't send because of problems", slog.Attr{Key: "msg.Id", Value: slog.StringValue(key)}, slog.Attr{Key: "Value", Value: slog.StringValue(string(jsonData))})
			time.Sleep(1 * time.Second)
			continue
		}
		resp.Body.Close()
		return
	}
}
