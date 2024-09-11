package hypertext

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/alexandernizov/checkChat/domain"
	"github.com/google/uuid"
)

type message struct {
	Id         int
	AuthorUuid uuid.UUID
	Body       string
	Published  time.Time
}

type Hypertext struct {
	log          *slog.Logger
	chatMessages <-chan domain.KafkaMessage
}

func New(log *slog.Logger, chatMessages <-chan domain.KafkaMessage) *Hypertext {
	return &Hypertext{log: log, chatMessages: chatMessages}
}

func (h *Hypertext) Serve() {
	go h.logHypertextFromMessage()
}

func (h *Hypertext) logHypertextFromMessage() {
	re := regexp.MustCompile(`https?://[^\s]+`)
	for msg := range h.chatMessages {
		id, body := getBodyFromKafkaMessage(msg)
		urls := re.FindAllString(body, -1)
		for _, url := range urls {
			h.log.Info("new urls", slog.Attr{Key: "msg.Id", Value: slog.IntValue(id)}, slog.Attr{Key: "url", Value: slog.StringValue(url)})
		}
	}
}

func getBodyFromKafkaMessage(msg domain.KafkaMessage) (int, string) {
	var message message
	err := json.Unmarshal([]byte(msg.Value), &message)
	if err != nil {
		fmt.Println("something gone wrong")
	}
	return message.Id, message.Body
}
