package botconn

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

var inputSample = `
{
    "message": {
        "data": {
            "userName": "Yannick Heinrich"
        },
        "__v": 0,
        "participant": "c9244b31-45f2-431c-be10-d3361851cf7e",
        "conversation": "f206b482-cb0c-435b-91bc-4628c8829d83",
        "attachment": {
            "content": "Hello",
            "type": "text"
        },
        "receivedAt": "2017-03-20T21:58:52.346Z",
        "isActive": true,
        "_id": "61a7921b-f771-4211-82ca-05885160fd6d"
    },
    "chatId": "chatId",
    "senderId": "senderId"
}`

func TestDecodeInputMessage(t *testing.T) {

	m := InputMessage{}
	err := json.Unmarshal([]byte(inputSample), &m)
	if err != nil {
		t.Errorf("Should have been able to parse the sample: %+v \n", err)
	}

	if m.ChatID != "chatId" {
		t.Errorf("Should decode ChatID")
	}

	if m.SenderID != "senderId" {
		t.Errorf("Should decode SenderID")
	}

	if m.Participant != "c9244b31-45f2-431c-be10-d3361851cf7e" {
		t.Errorf("Should decode Participant")
	}

	if m.Conversation != "f206b482-cb0c-435b-91bc-4628c8829d83" {
		t.Errorf("Should decode Conversation")

	}

	validTime, _ := time.Parse(time.RFC3339, "2017-03-20T21:58:52.346Z")
	if validTime != m.Received {
		t.Errorf("Should decode Received")
	}

	validAttachment := Attachment{
		Content: "Hello",
		Kind:    TextKind,
	}

	if validAttachment != m.Attachment {
		t.Errorf("Should decode Attachment")
	}

	if value, ok := m.Data.(map[string]interface{}); !ok {
		t.Errorf("Should decode Data")
	} else {
		if value["userName"] != "Yannick Heinrich" {
			t.Errorf("Should decode Data: %+v \n", value)
		}
	}
}

func TestEncodeOutputMessage(t *testing.T) {
	output := OutputMessage{
		Kind:    TextKind,
		Content: "Coucou",
	}
	value, err := json.Marshal(output)
	if err != nil {
		t.Errorf("Should have succeeded: %s \n", err)
		t.FailNow()
	}

	var unmarshalled map[string]string

	err = json.Unmarshal(value, &unmarshalled)
	if err != nil {
		t.FailNow()
	}

	expectedValue := map[string]string{
		"content": "Coucou",
		"type":    "text",
	}

	if !reflect.DeepEqual(expectedValue, unmarshalled) {
		t.Errorf("Should be the same value - Expected: %+v | Computed: %+v \n", expectedValue, unmarshalled)
	}
}
