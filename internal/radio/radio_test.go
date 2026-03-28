package radio

import (
	"testing"
	"time"
)

func TestNewLogIsEmpty(t *testing.T) {
	log := NewLog()
	if log.Len() != 0 {
		t.Errorf("expected empty log, got %d messages", log.Len())
	}
}

func TestAddMessage(t *testing.T) {
	log := NewLog()
	msg := PilotMessage(10*time.Second, "AA123", "requesting approach")

	log2 := log.Add(msg)

	if log.Len() != 0 {
		t.Error("original log should be unchanged (immutability)")
	}
	if log2.Len() != 1 {
		t.Errorf("expected 1 message, got %d", log2.Len())
	}
}

func TestAddPreservesOrder(t *testing.T) {
	log := NewLog()
	log = log.Add(PilotMessage(1*time.Second, "AA1", "first"))
	log = log.Add(ATCMessage(2*time.Second, "AA1", "second"))
	log = log.Add(PilotMessage(3*time.Second, "AA2", "third"))

	msgs := log.All()
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].Text != "first" || msgs[1].Text != "second" || msgs[2].Text != "third" {
		t.Error("messages not in expected order")
	}
}

func TestAddTrimsOldest(t *testing.T) {
	log := NewLog()
	for i := 0; i < MaxLogSize+10; i++ {
		log = log.Add(PilotMessage(time.Duration(i)*time.Second, "AA1", "msg"))
	}

	if log.Len() != MaxLogSize {
		t.Errorf("expected %d messages after trim, got %d", MaxLogSize, log.Len())
	}

	// Oldest should have been trimmed — first message time should be 10s
	msgs := log.All()
	if msgs[0].Time != 10*time.Second {
		t.Errorf("expected oldest message at 10s, got %v", msgs[0].Time)
	}
}

func TestLastN(t *testing.T) {
	log := NewLog()
	log = log.Add(PilotMessage(1*time.Second, "AA1", "old"))
	log = log.Add(ATCMessage(2*time.Second, "AA1", "mid"))
	log = log.Add(PilotMessage(3*time.Second, "AA1", "new"))

	last2 := log.Last(2)
	if len(last2) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(last2))
	}
	if last2[0].Text != "mid" || last2[1].Text != "new" {
		t.Error("Last(2) returned wrong messages")
	}
}

func TestLastNExceedsLen(t *testing.T) {
	log := NewLog()
	log = log.Add(PilotMessage(1*time.Second, "AA1", "only"))

	last5 := log.Last(5)
	if len(last5) != 1 {
		t.Errorf("expected 1 message, got %d", len(last5))
	}
}

func TestPilotMessage(t *testing.T) {
	msg := PilotMessage(30*time.Second, "UA456", "with you at 5000")
	if msg.Direction != Inbound {
		t.Error("pilot message should be inbound")
	}
	if msg.From != "UA456" || msg.To != "ATC" {
		t.Errorf("expected from=UA456 to=ATC, got from=%s to=%s", msg.From, msg.To)
	}
	if msg.Priority != Normal {
		t.Error("pilot message should default to normal priority")
	}
}

func TestATCMessage(t *testing.T) {
	msg := ATCMessage(45*time.Second, "DL789", "turn right heading 090")
	if msg.Direction != Outbound {
		t.Error("ATC message should be outbound")
	}
	if msg.From != "ATC" || msg.To != "DL789" {
		t.Errorf("expected from=ATC to=DL789, got from=%s to=%s", msg.From, msg.To)
	}
}

func TestSystemMessage(t *testing.T) {
	msg := SystemMessage(60*time.Second, "COLLISION WARNING", Emergency)
	if msg.Priority != Emergency {
		t.Errorf("expected emergency priority, got %v", msg.Priority)
	}
	if msg.From != "SYS" {
		t.Errorf("expected from=SYS, got %s", msg.From)
	}
}

func TestDirectionString(t *testing.T) {
	tests := []struct {
		dir  Direction
		want string
	}{
		{Inbound, "IN"},
		{Outbound, "OUT"},
		{Direction(99), "?"},
	}
	for _, tt := range tests {
		if got := tt.dir.String(); got != tt.want {
			t.Errorf("Direction(%d).String() = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		pri  Priority
		want string
	}{
		{Normal, "NORMAL"},
		{Urgent, "URGENT"},
		{Emergency, "EMERGENCY"},
		{Priority(99), "?"},
	}
	for _, tt := range tests {
		if got := tt.pri.String(); got != tt.want {
			t.Errorf("Priority(%d).String() = %q, want %q", tt.pri, got, tt.want)
		}
	}
}

func TestImmutability(t *testing.T) {
	log1 := NewLog()
	log1 = log1.Add(PilotMessage(1*time.Second, "AA1", "hello"))

	log2 := log1.Add(ATCMessage(2*time.Second, "AA1", "roger"))

	if log1.Len() != 1 {
		t.Errorf("log1 should still have 1 message, got %d", log1.Len())
	}
	if log2.Len() != 2 {
		t.Errorf("log2 should have 2 messages, got %d", log2.Len())
	}

	// Mutating the returned slice should not affect the log
	msgs := log1.All()
	msgs[0].Text = "mutated"
	if log1.All()[0].Text == "mutated" {
		t.Error("modifying returned slice should not affect the log")
	}
}
