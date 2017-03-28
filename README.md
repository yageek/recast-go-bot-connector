[![GoDoc](https://godoc.org/github.com/yageek/recast-go-bot-connector?status.png)](https://godoc.org/github.com/yageek/recast-go-bot-connector) 
[![Report Cart](http://goreportcard.com/badge/yageek/recast-go-bot-connector)](http://goreportcard.com/report/yageek/recast-go-bot-connector)
# recast-go-bot-connector

Package helping to deal with the [Recast.AI Bot Connector](https://botconnector.recast.ai/)

## Installation

```
go get -u github.com/yageek/recast-go-bot-connector
```

## Usage
### Catch and reply

```go
package main

import (
    "fmt"
    "github.com/bmizerany/pat"
    "github.com/yageek/recast-go-bot-connector"
)
func main() {

    conf := botconn.ConnConfig{
		Domain:    botconn.RecastAPIDomain,
		BotID:     "BOT_ID",
		UserSlug:  "USER_SLUG",
		UserToken: "USER_TOKEN",
	}
	conn := botconn.New(conf)
	// Message routing
	conn.UseHandler(botconn.MessageHandlerFunc(nextBus))

	// Router
	mux := pat.New()
	mux.Post("/chatbot", conn)

	http.HandleFunc("/", mux)
    http.ListenAndServe(":8080", nil)
}

func nextBus(m botconn.InputMessage, c botconn.ConnConfig) {

	fmt.Printf("Message: %+v config: %+v \n", m, c)
}

```
### Push a message to one participant

```go
    output := botconn.OutputMessage{
		Content: "Coucou",
		Kind:    botconn.TextKind,
	}
    err := conn.Send(output, "CONVERSATION_ID", "SENDER_ID")
    if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Response succeeded")
	}
```

### Broadcast a message to all participants

```go
    output := botconn.OutputMessage{
		Content: "Coucou",
		Kind:    botconn.TextKind,
	}
    err := conn.Broadcast(output)
    if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Response succeeded")
	}
```