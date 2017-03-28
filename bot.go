package botconn

import (
	"encoding/json"
	"net/http"
)

var (
	RecastAPIDomain = "https://api-botconnector.recast.ai"
)

// MessageHandler is implemented
// by any element that would like to receive a message.
type MessageHandler interface {
	ServeMessage(m InputMessage, config ConnConfig)
}

type MessageHandlerFunc func(m InputMessage, config ConnConfig)

func (f MessageHandlerFunc) ServeMessage(m InputMessage, config ConnConfig) {
	f(m, config)
}

// ConnConfig is used to configure
// the connector.
type ConnConfig struct {
	Domain    string
	BotID     string
	UserSlug  string
	UserToken string
}

// Connector is a connector
// to the recast connnector API.
type Connector struct {
	config  ConnConfig
	handler MessageHandler
}

// New creates a new connector with
// the provided configuration
func New(c ConnConfig) *Connector {
	return &Connector{
		config: c,
	}
}

// UseHandler specifies the
// receiver for the message.
func (c *Connector) UseHandler(h MessageHandler) {
	c.handler = h
}

func (c *Connector) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	message := InputMessage{}
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid Content:", http.StatusBadRequest)
		return
	}

	if c.handler != nil {
		c.handler.ServeMessage(message, c.config)
	}

	w.WriteHeader(http.StatusOK)
}
