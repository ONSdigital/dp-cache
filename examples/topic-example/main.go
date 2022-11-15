package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	topic "github.com/ONSdigital/dp-cache/examples/topic-example/topics"
)

func handlerOne(w http.ResponseWriter, r *http.Request) {}

const (
	PORT  = 4242
	route = "/topic"
)

func main() {
	uri := fmt.Sprintf("http://127.0.0.1:%d", PORT)
	log.Printf("Starting server on %s\n", uri)

	// ------------------------------------
	// Step #2
	// ------------------------------------
	ctx := context.Background()
	errChan := make(chan error)
	interval := 5 * time.Second
	// Step #2a
	topicCache, _ := topic.NewTopicCache(ctx, &interval)
	// Step #2b
	topicCache.AddUpdateFunc("main_topic", topic.UpdateTopic())
	// Step #2c
	go topicCache.StartUpdates(ctx, errChan)

	// Mux server
	server := http.Server{
		Addr: fmt.Sprintf(":%d", PORT),
	}
	// Middleware
	middleware := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			topic, _ := topicCache.Get("main_topic")
			fmt.Fprintf(w, "%+v\n", topic)
			h(w, r)
		}
	}

	// Handlers
	http.HandleFunc(route, middleware(handlerOne))

	log.Printf("Hit the following  %s%s\n", uri, route)

	// Start server
	err := server.ListenAndServe()
	if err != nil {
		// Step #2d
		topicCache.Close()
	}
}
