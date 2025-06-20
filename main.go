package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Config struct {
	Url      string
	Stream   string
	Subjects []string
	// Much more should go here, but I just need a simple version of this for now.
}

type NatsContext struct {
	ContextName string
	Url         string `json:"url"`
}

var query string
var stream string
var subjects string

func main() {
	processFlags()
	godotenv.Load()
	conf := loadConfig()

	nc, err := nats.Connect(conf.Url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to NATS: %v\n", err)
		panic("Failed to connect to NATS")
	}
	defer nc.Close()
	js, err := jetstream.New(nc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating JetStream context: %v\n", err)
		panic("Failed to create JetStream context")
	}
	consumer, err := js.CreateConsumer(
		context.Background(),
		conf.Stream,
		jetstream.ConsumerConfig{
			FilterSubjects: conf.Subjects,
		},
	)

	for {
		msg, err := consumer.Fetch(100, jetstream.FetchMaxWait(3*time.Second))
		if err != nil {
			panic(err)
		}

		ctr := 0
		for m := range msg.Messages() {
			ctr++
			data := string(m.Data())
			if query == "" || strings.Contains(data, query) {
				fmt.Println(string(m.Data()))
			}
		}
		if ctr == 0 {
			fmt.Println("No messages received, exiting.")
			break
		}
	}
	os.Exit(0)
}

// Load CLI context, if able.
func loadConfig() Config {
	natsContext := loadNatsContext()
	subjectList := strings.Split(subjects, ",")

	return Config{
		Url:      natsContext.Url,
		Stream:   stream,
		Subjects: subjectList,
	}
}

func loadNatsContext() NatsContext {
	// TODO: Add some other way to populate this. Also, there's probably some set of
	// paths we should check for the context file.
	ctxName, err := readFile(os.ExpandEnv("$HOME/.config/nats/context.txt"))
	if err != nil {
		// Shouldn't error out, eventually
		fmt.Fprintf(os.Stderr, "Error reading context file: %v\n", err)
		panic("Failed to read context file")
	}

	natsCtx, err := readContextJSONFile(os.ExpandEnv("$HOME/.config/nats/context/" + ctxName + ".json"))
	if err != nil {
		// Once again, we'll change this later.
		panic(fmt.Sprintf("Failed to read context JSON file: %v", err))
	}

	natsCtx.ContextName = ctxName

	return natsCtx
}

func readFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return string(data), nil
}

func readContextJSONFile(filename string) (NatsContext, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return NatsContext{}, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	var ctx NatsContext
	err = json.Unmarshal(data, &ctx)
	if err != nil {
		return NatsContext{}, fmt.Errorf("failed to unmarshal JSON from file %s: %w", filename, err)
	}
	return ctx, nil
}

func processFlags() {
	flag.StringVar(&query, "query", "", "Query to filter messages")
	flag.StringVar(&stream, "stream", "", "JetStream stream name. Required")
	flag.StringVar(&subjects, "subjects", "", "Subjects to filter messages. Comma-separated list with no spaces. Required")
	flag.Parse()

	if stream == "" {
		fmt.Println("No stream provided, exiting.")
		os.Exit(1)
	}

	if subjects == "" {
		fmt.Println("No subjects provided, exiting.")
		os.Exit(1)
	}
}
