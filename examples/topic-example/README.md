# Example (Topic) Caching Server

### Install & Start server

Option 1: From root of repository, run: `make topic-example` 
Option 2: From topic-example directory, run: `HUMAN_LOG=1 go run -race examples/topic-example/main.go`

Navigate to [http://127.0.0.1:4242/topic](http://127.0.0.1:4242/topic)
Page will update the topic ID every 5 seconds.