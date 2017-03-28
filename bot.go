package botconn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var (
	RecastAPIDomain    = "https://api-botconnector.recast.ai"
	ErrResponseNotSent = errors.New("Message not sent")
)

// MessageHandler is implemented
// by any element that would like to receive a message.
type MessageHandler interface {
	ServeMessage(w MessageWriter, m InputMessage)
}

// MessageWriter is the structure
// allowing you to respond.
type MessageWriter interface {
	Reply(o OutputMessage) error
}

type outMessagePayload struct {
	Messages []OutputMessage `json:"messages"`
	SenderID string          `json:"senderId"`
}

type writer struct {
	config         ConnConfig
	conversationID string
	senderID       string
	client         *http.Client
}

func newWriter(conversationID, senderID string, c ConnConfig, client *http.Client) *writer {
	return &writer{
		conversationID: conversationID,
		senderID:       senderID,
		config:         c,
		client:         client,
	}
}
func (w *writer) replyURL() (*url.URL, error) {

	c := w.config
	rawString := fmt.Sprintf("%s/users/%s/bots/%s/conversations/%s/messages", c.Domain, c.UserSlug, c.BotID, w.conversationID)
	return url.Parse(rawString)
}
func (w *writer) Reply(message OutputMessage) error {
	return w.ReplyMultiple([]OutputMessage{message})
}

func (w *writer) ReplyMultiple(messages []OutputMessage) error {
	// Try marshalling
	buff := new(bytes.Buffer)

	out := outMessagePayload{
		Messages: messages,
		SenderID: w.senderID,
	}
	err := json.NewEncoder(buff).Encode(out)
	if err != nil {
		return err
	}

	replyURL, err := w.replyURL()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", replyURL.String(), buff)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", w.config.UserToken))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := w.client.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return ErrResponseNotSent
	}
	return err
}

func (w *writer) String() string {
	return fmt.Sprintf("MessageWriter<%p>", w)
}

// MessageHandlerFunc simpler wrapper for function.
type MessageHandlerFunc func(w MessageWriter, m InputMessage)

func (f MessageHandlerFunc) ServeMessage(w MessageWriter, m InputMessage) {
	f(w, m)
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
	client  *http.Client
}

// New creates a new connector with
// the provided configuration
func New(c ConnConfig) *Connector {
	return &Connector{
		config: c,
		client: &http.Client{},
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
		w := newWriter(message.Conversation, message.SenderID, c.config, c.client)
		go c.handler.ServeMessage(w, message)
	}

	w.WriteHeader(http.StatusOK)
}
