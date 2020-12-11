package protocol

const (
	/**
	Message with connection options
	*/
	MessageTypeOpen = iota
	/**
	Close connection and destroy all handle routines
	*/
	MessageTypeClose
	/**
	Ping request message
	*/
	MessageTypePing
	/**
	Pong response message
	*/
	MessageTypePong
	/**
	Empty message
	*/
	MessageTypeEmpty
	/**
	Emit request, no response
	*/
	MessageTypeEmit
	/**
	Emit request, wait for response (ack)
	*/
	MessageTypeAckRequest
	/**
	ack response
	*/
	MessageTypeAckResponse
)

type Message struct {
	Type   int
	AckId  int
	Method string
	Args   interface{}
	Source string
}

