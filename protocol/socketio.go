package protocol

import (
	"encoding/json"
	"errors"
	"github.com/valyala/fastjson"
	"strconv"
	"strings"
)

const (
	open          = "0"
	msg           = "4"
	emptyMessage  = "40"
	commonMessage = "42"
	ackMessage    = "43"

	CloseMessage = "1"
	PingMessage = "2"
	PongMessage = "3"
)

var (
	ErrorWrongMessageType = errors.New("Wrong message type")
	ErrorWrongPacket      = errors.New("Wrong packet")
)

func typeToText(msgType int) (string, error) {
	switch msgType {
	case MessageTypeOpen:
		return open, nil
	case MessageTypeClose:
		return CloseMessage, nil
	case MessageTypePing:
		return PingMessage, nil
	case MessageTypePong:
		return PongMessage, nil
	case MessageTypeEmpty:
		return emptyMessage, nil
	case MessageTypeEmit, MessageTypeAckRequest:
		return commonMessage, nil
	case MessageTypeAckResponse:
		return ackMessage, nil
	}
	return "", ErrorWrongMessageType
}

func Encode(msg *Message) (string, error) {
	result, err := typeToText(msg.Type)
	if err != nil {
		return "", err
	}

	if msg.Type == MessageTypeEmpty || msg.Type == MessageTypePing ||
		msg.Type == MessageTypePong {
		return result, nil
	}

	if msg.Type == MessageTypeAckRequest || msg.Type == MessageTypeAckResponse {
		result += strconv.Itoa(msg.AckId)
	}
	var argStr string

	if msg.Args != nil {
		argVal, err := json.Marshal(msg.Args)

		if err != nil {
			return "", err
		}

		argStr = string(argVal)
	}

	if msg.Type == MessageTypeOpen || msg.Type == MessageTypeClose {
		return result + argStr, nil
	}

	if msg.Type == MessageTypeAckResponse {
		return result + argStr, nil
	}

	jsonMethod, err := json.Marshal(&msg.Method)

	if err != nil {
		return "", err
	}

	if msg.Args == nil {
		return result + "[" + string(jsonMethod) + "]", nil
	}

	var ok bool
	var args []interface{}

	if args, ok = msg.Args.([]interface{}); !ok {
		args = []interface{}{msg.Args}
	}

	msg.Args = append([]interface{}{msg.Method}, args...)

	argVal, err := json.Marshal(msg.Args)

	if err != nil {
		return "", err
	}

	argStr = string(argVal)

	return result + argStr, nil
}

func MustEncode(msg *Message) string {
	result, err := Encode(msg)
	if err != nil {
		panic(err)
	}

	return result
}

func getMessageType(data string) (int, error) {
	if len(data) == 0 {
		return 0, ErrorWrongMessageType
	}
	switch data[0:1] {
	case open:
		return MessageTypeOpen, nil
	case CloseMessage:
		return MessageTypeClose, nil
	case PingMessage:
		return MessageTypePing, nil
	case PongMessage:
		return MessageTypePong, nil
	case msg:
		if len(data) == 1 {
			return 0, ErrorWrongMessageType
		}
		switch data[0:2] {
		case emptyMessage:
			return MessageTypeEmpty, nil
		case commonMessage:
			return MessageTypeAckRequest, nil
		case ackMessage:
			return MessageTypeAckResponse, nil
		}
	}
	return 0, ErrorWrongMessageType
}

/**
Get ack id of current packet, if present
*/
func getAck(text string) (ackId int, restText string, err error) {
	if len(text) < 4 {
		return 0, "", ErrorWrongPacket
	}
	text = text[2:]

	pos := strings.IndexByte(text, '[')
	if pos == -1 {
		return 0, "", ErrorWrongPacket
	}

	ack, err := strconv.Atoi(text[0:pos])
	if err != nil {
		return 0, "", err
	}

	return ack, text[pos:], nil
}

var (
	pool fastjson.ParserPool
)

func Decode(data string) (*Message, error) {
	var err error
	msg := &Message{}
	msg.Source = data

	msg.Type, err = getMessageType(data)
	if err != nil {
		return nil, err
	}

	if msg.Type == MessageTypeClose || msg.Type == MessageTypePing ||
		msg.Type == MessageTypePong || msg.Type == MessageTypeEmpty {
		return msg, nil
	}

	if msg.Type == MessageTypeOpen {
		vals := make([]interface{}, 0)

		if err = json.Unmarshal([]byte(data[1:]), &vals); err != nil {
			return nil, err
		}

		msg.Args = vals
		return msg, nil
	}

	ack, rest, err := getAck(data)
	msg.AckId = ack
	if msg.Type == MessageTypeAckResponse {
		if err != nil {
			return nil, err
		}
		vals := make([]interface{}, 0)

		if err = json.Unmarshal([]byte(rest), &vals); err != nil {
			return nil, err
		}

		msg.Args = vals
		return msg, nil
	}

	if err != nil {
		msg.Type = MessageTypeEmit
		rest = data[2:]
	}

	vals := make([]interface{}, 0)

	if err = json.Unmarshal([]byte(rest), &vals); err != nil {
		return nil, err
	}

	msg.Method = vals[0].(string)

	msg.Args = vals[1:]

	return msg, nil
}
