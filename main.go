package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/alexandernizov/checkChat/consumer"
	"github.com/alexandernizov/checkChat/hypertext"
	"github.com/alexandernizov/checkChat/webhook"
	"github.com/ilyakaznacheev/cleanenv"
)

func main() {
	cfg := MustLoad()

	file, err := os.OpenFile(cfg.Log.Filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("Failed to open log file %w", err)
		os.Exit(1)
	}
	defer file.Close()

	log := slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{Level: slog.LevelDebug}))

	chatConsumer, err := consumer.New(cfg.СhatsConsumer.Topic, cfg.СhatsConsumer.Host, cfg.СhatsConsumer.GroupId)
	if err != nil {
		fmt.Println("error to start consumer", err)
	}
	defer chatConsumer.Stop()
	chatConsumer.Consume()

	webhook := webhook.New(cfg.СhatsConsumer.Url, chatConsumer.ResultChannel(), cfg.СhatsConsumer.Retry)
	webhook.Serve()

	messagesConsumer, err := consumer.New(cfg.MessagesConsumer.Topic, cfg.MessagesConsumer.Host, cfg.MessagesConsumer.GroupId)
	if err != nil {
		fmt.Println("error to start consumer", err)
	}
	defer messagesConsumer.Stop()
	messagesConsumer.Consume()

	hyperText := hypertext.New(log, messagesConsumer.ResultChannel())

	hyperText.Serve()

	time.Sleep(1 * time.Hour)
}

type Config struct {
	Env              string                 `yaml:"env"`
	Log              LogConfig              `yaml:"log"`
	СhatsConsumer    ChatsConsumerConfig    `yaml:"chatsConsumer"`
	MessagesConsumer MessagesConsumerConfig `yaml:"messagesConsumer"`
}

type LogConfig struct {
	Filename string `yaml:"filename"`
}

type ChatsConsumerConfig struct {
	Host    string `yaml:"host"`
	Topic   string `yaml:"topic"`
	GroupId string `yaml:"groupId"`
	Url     string `yaml:"url"`
	Retry   int    `yaml:"retry"`
}

type MessagesConsumerConfig struct {
	Host    string `yaml:"host"`
	Topic   string `yaml:"topic"`
	GroupId string `yaml:"groupId"`
}

func MustLoad() *Config {
	path := fetchFlags()
	if path == "" {
		path = "./config.yml"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists: " + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}
	return &cfg
}

func fetchFlags() string {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	return configPath
}
