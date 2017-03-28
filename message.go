package botconn

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidContent represent a generic unmarshalling error.
	ErrInvalidContent = errors.New("The provided content is invalid.")
	ErrInvalidKind    = errors.New("The provided content type is invalid.")
)

// ContentKind represents a kind of content
type ContentKind int

const (
	// TextKind represents a plain text content
	TextKind ContentKind = iota
	// TextKind represents an unknown content
	UnknownContent
)

// Attachment contains the content of the message
type Attachment struct {
	Content string      `json:"content"`
	Kind    ContentKind `json:"type"`
}

// InputMessage is a message received
type InputMessage struct {
	Data         interface{}
	SenderID     string
	ChatID       string
	Participant  string
	Conversation string
	Attachment   Attachment
	Received     time.Time
}

func getElement(k string, values map[string]interface{}) (interface{}, error) {
	v, ok := values[k]
	if !ok {
		return nil, fmt.Errorf("No %s key was found \n", k)
	}
	return v, nil
}
func getString(k string, values map[string]interface{}) (string, error) {
	v, err := getElement(k, values)
	if err != nil {
		return "", err
	}
	if value, ok := v.(string); !ok {
		return "", ErrInvalidContent
	} else {
		return value, nil
	}
}

func getAttachment(k string, values map[string]interface{}) (Attachment, error) {
	v, err := getElement(k, values)
	if err != nil {
		return Attachment{}, err
	}

	rawAttachment, ok := v.(map[string]interface{})
	if !ok {
		return Attachment{}, ErrInvalidContent
	}

	typeAttachment, err := getString("type", rawAttachment)
	if err != nil {
		return Attachment{}, ErrInvalidContent
	}
	content, err := getString("content", rawAttachment)
	if err != nil {
		return Attachment{}, ErrInvalidContent
	}

	kind, err := getKind(typeAttachment)
	if err != nil {
		return Attachment{}, err
	}

	return Attachment{
		Kind:    kind,
		Content: content,
	}, nil
}

func getKind(k string) (ContentKind, error) {
	switch k {
	case "text":
		return TextKind, nil
	default:
		return UnknownContent, ErrInvalidKind
	}

}
func getKindString(k ContentKind) (string, error) {
	switch k {
	case TextKind:
		return "text", nil
	default:
		return "", ErrInvalidKind
	}

}

// UnmarshalJSON allows custom unmarshalling
func (i *InputMessage) UnmarshalJSON(data []byte) error {
	var dst map[string]interface{}
	err := json.Unmarshal(data, &dst)

	if err != nil {
		return err
	}
	_, ok := dst["message"]
	if !ok {
		return ErrInvalidContent
	}

	rawMessage, ok := dst["message"].(map[string]interface{})
	if !ok {
		return ErrInvalidContent
	}

	// SenderID
	if value, err := getString("senderId", dst); err != nil {
		return err
	} else {
		i.SenderID = value
	}

	// ChatID
	if value, err := getString("chatId", dst); err != nil {
		return err
	} else {
		i.ChatID = value
	}

	// Participant
	if value, err := getString("participant", rawMessage); err != nil {
		return err
	} else {
		i.Participant = value
	}

	// Conversation
	if value, err := getString("conversation", rawMessage); err != nil {
		return err
	} else {
		i.Conversation = value
	}

	//Time
	if value, err := getString("receivedAt", rawMessage); err != nil {
		return err
	} else {
		t, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return err
		}
		i.Received = t
	}
	//Attachment
	if value, err := getAttachment("attachment", rawMessage); err != nil {
		return err
	} else {
		i.Attachment = value
	}
	// Data
	if value, ok := rawMessage["data"]; ok {
		i.Data = value
	}
	return nil
}

// OutputMessage is a message sent to from the bot.
type OutputMessage struct {
	Kind    ContentKind
	Content string
}

func (o OutputMessage) MarshalJSON() ([]byte, error) {

	kindString, err := getKindString(o.Kind)
	if err != nil {
		return []byte{}, err
	}
	rawOutput := map[string]string{
		"type":    kindString,
		"content": o.Content,
	}
	return json.Marshal(rawOutput)
}
