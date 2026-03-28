package radio

import "time"

// Direction indicates whether a message is inbound (pilot→ATC) or outbound (ATC→pilot).
type Direction int

const (
	Inbound  Direction = iota // Pilot → ATC
	Outbound                  // ATC → Pilot
)

func (d Direction) String() string {
	switch d {
	case Inbound:
		return "IN"
	case Outbound:
		return "OUT"
	default:
		return "?"
	}
}

// Priority controls message styling urgency.
type Priority int

const (
	Normal    Priority = iota // Standard comms
	Urgent                    // Repeated request, "say again"
	Emergency                 // Collision warning, go-around
)

func (p Priority) String() string {
	switch p {
	case Normal:
		return "NORMAL"
	case Urgent:
		return "URGENT"
	case Emergency:
		return "EMERGENCY"
	default:
		return "?"
	}
}

// Message represents a single radio communication.
type Message struct {
	Time      time.Duration
	From      string
	To        string
	Text      string
	Direction Direction
	Priority  Priority
}

// MaxLogSize is the maximum number of messages retained in the log.
const MaxLogSize = 100

// Log is an append-only radio communications log.
type Log struct {
	messages []Message
}

// NewLog creates an empty radio log.
func NewLog() Log {
	return Log{
		messages: make([]Message, 0, MaxLogSize),
	}
}

// Add appends a message to the log, returning a new Log. Trims oldest messages
// if the log exceeds MaxLogSize.
func (l Log) Add(msg Message) Log {
	msgs := make([]Message, len(l.messages), len(l.messages)+1)
	copy(msgs, l.messages)
	msgs = append(msgs, msg)
	if len(msgs) > MaxLogSize {
		msgs = msgs[len(msgs)-MaxLogSize:]
	}
	return Log{messages: msgs}
}

// All returns all messages in the log.
func (l Log) All() []Message {
	result := make([]Message, len(l.messages))
	copy(result, l.messages)
	return result
}

// Last returns the most recent n messages.
func (l Log) Last(n int) []Message {
	if n >= len(l.messages) {
		return l.All()
	}
	start := len(l.messages) - n
	result := make([]Message, n)
	copy(result, l.messages[start:])
	return result
}

// Len returns the number of messages in the log.
func (l Log) Len() int {
	return len(l.messages)
}

// PilotMessage creates an inbound message from a pilot to ATC.
func PilotMessage(elapsed time.Duration, callsign, text string) Message {
	return Message{
		Time:      elapsed,
		From:      callsign,
		To:        "ATC",
		Text:      text,
		Direction: Inbound,
		Priority:  Normal,
	}
}

// ATCMessage creates an outbound message from ATC to a pilot.
func ATCMessage(elapsed time.Duration, callsign, text string) Message {
	return Message{
		Time:      elapsed,
		From:      "ATC",
		To:        callsign,
		Text:      text,
		Direction: Outbound,
		Priority:  Normal,
	}
}

// SystemMessage creates an emergency/system message (e.g., collision warning).
func SystemMessage(elapsed time.Duration, text string, priority Priority) Message {
	return Message{
		Time:      elapsed,
		From:      "SYS",
		To:        "",
		Text:      text,
		Direction: Inbound,
		Priority:  priority,
	}
}
