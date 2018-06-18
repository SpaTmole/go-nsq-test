package main

import (
	"errors"
	"fmt"
	"github.com/nsqio/go-nsq"
	"math/rand"
	"os"
	"time"
)

const CHANNEL = "TestChann"
const TOPIC = "TestTopic"
const USE_DEFLATE = false
const USE_SNAPPY = true
const DEFLATE_LVL = 3
const KB = 1024
const MB = 1024 * KB

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Queue struct {
	ch chan string
}

type ExchangeHandler struct {
	Q *Queue
}

func (q *Queue) Put(value string) {
	q.ch <- value
}

func (q *Queue) Get(timeout int64) (res string, err error) {
	select {
	case res = <-q.ch:
		return
	case <-time.After(time.Duration(timeout) * time.Second):
		err = errors.New("Timeout.")
		return
	}
}

func (h *ExchangeHandler) HandleMessage(message *nsq.Message) error {
	res := string(message.Body)
	message.Finish()
	h.Q.Put(res)
	return nil
}

func (h *ExchangeHandler) ReadMessage() (res string, err error) {
	res, err = h.Q.Get(10)
	return
}

func NewQueue() (res *Queue) {
	res = &Queue{ch: make(chan string, 1)}
	return
}

func NewExchangeHandler() (res *ExchangeHandler) {
	res = &ExchangeHandler{Q: NewQueue()}
	return
}

func config() *nsq.Config {
	cfg := nsq.NewConfig()
	cfg.Deflate = USE_DEFLATE
	cfg.DeflateLevel = DEFLATE_LVL
	cfg.Snappy = USE_SNAPPY
	if err := cfg.Validate(); err != nil {
		fmt.Println("Config is incorrect: ", err)
		return nil
	}
	return cfg
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	var message string
	var err error
	sizes := []int{
		KB,
		10 * KB,
		60 * KB,
		500 * KB,
		MB,
		2 * MB,
		5 * MB,
		10 * MB,
		90 * MB,   // Started failing in my case.
		1000 * MB, // SNAPPY needed much bigger messages.
	}
	consumer, err := nsq.NewConsumer(TOPIC, CHANNEL, config())
	if err != nil {
		fmt.Println("Couldn't create consumer: ", err)
		os.Exit(1)
	}
	handler := NewExchangeHandler()
	consumer.AddHandler(handler)
	err = consumer.ConnectToNSQD("nsqd:4150")
	if err != nil {
		fmt.Println("Couldn't connect to consumer's lookup: ", err)
		os.Exit(1)
	}
	producer, err := nsq.NewProducer("nsqd:4150", config())
	if err != nil {
		fmt.Println("Couldn't create publisher: ", err)
		os.Exit(1)
	}
	for _, size := range sizes {
		message = randSeq(size)
		err = producer.PublishAsync(TOPIC, []byte(message), nil)
		if err != nil {
			fmt.Println("Couldn't publish: ", err)
			os.Exit(1)
		}
		res, err := handler.ReadMessage()
		fmt.Println(bool(res == message && err == nil), err, size)
	}
	consumer.Stop()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
