# go-bo-connector

Port of https://github.com/RecastAI/SDK-NodeJS-bot-connector.

## Installation

```
go get -u github.com/yageek/recast-go-bot-connector
```

## Usage

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