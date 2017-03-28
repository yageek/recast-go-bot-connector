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
	RecastAPIDomain      = "https://api-botconnector.recast.ai"
	ErrResponseNotSent   = errors.New("Message not sent")
	ErrInvalidStatusCode = errors.New("invalid status code")
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
	SenderID string          `json:"senderId,omitempty"`
}

type writer struct {
	config         ConnConfig
	conversationID string
	senderID       string
	client         *http.Client
}

func replyURL(domain, userSlug, botID, conversationID string) (*url.URL, error) {
	rawString := fmt.Sprintf("%s/users/%s/bots/%s/conversations/%s/messages", domain, userSlug, botID, conversationID)
	return url.Parse(rawString)
}

func broadcastURL(domain, userSlug, botID string) (*url.URL, error) {
	rawString := fmt.Sprintf("%s/users/%s/bots/%s/messages", domain, userSlug, botID)
	return url.Parse(rawString)
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
	return replyURL(c.Domain, c.UserSlug, c.BotID, w.conversationID)
}
func (w *writer) Reply(message OutputMessage) error {
	return w.ReplyMultiple([]OutputMessage{message})
}

func sendJSON(client *http.Client, url *url.URL, v interface{}, statusCode int, token string) error {
	// Try marshalling
	buff := new(bytes.Buffer)

	err := json.NewEncoder(buff).Encode(v)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url.String(), buff)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)

	defer resp.Body.Close()

	if resp.StatusCode != statusCode {
		return ErrInvalidStatusCode
	}
	return err
}
func (w *writer) ReplyMultiple(messages []OutputMessage) error {

	replyURL, err := w.replyURL()
	if err != nil {
		return err
	}
	out := outMessagePayload{
		Messages: messages,
		SenderID: w.senderID,
	}
	err = sendJSON(w.client, replyURL, out, http.StatusCreated, w.config.UserToken)
	if err != nil {
		return ErrResponseNotSent
	}
	return nil
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

// Send a message
func (c *Connector) Send(message OutputMessage, conversationID, senderID string) error {

	sendURL, err := replyURL(c.config.Domain, c.config.UserSlug, c.config.BotID, conversationID)
	if err != nil {
		return err
	}
	out := outMessagePayload{
		Messages: []OutputMessage{message},
		SenderID: senderID,
	}
	return sendJSON(c.client, sendURL, out, http.StatusCreated, c.config.UserToken)
}

// Broadcast send a message to all participant.
func (c *Connector) Broadcast(message OutputMessage) error {
	broadCastURL, err := broadcastURL(c.config.Domain, c.config.UserSlug, c.config.BotID)
	if err != nil {
		return err
	}
	out := outMessagePayload{
		Messages: []OutputMessage{message},
	}

	return sendJSON(c.client, broadCastURL, out, http.StatusCreated, c.config.UserToken)

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
