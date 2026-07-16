package task

import (
	"encoding/json"
	"testing"
)

// TestNewSendEmailTask_PayloadRoundTrip is an assert-based smoke check (no
// framework): verifies the task type constant and that payload marshaling
// survives a round trip, since EmailHandler.ProcessTask depends on both.
func TestNewSendEmailTask_PayloadRoundTrip(t *testing.T) {
	in := SendEmailPayload{
		ToEmail:     "user@example.com",
		ToName:      "User",
		Subject:     "Update Pesanan #1",
		HTMLContent: "<p>hi</p>",
		TextContent: "hi",
	}

	task, err := NewSendEmailTask(in)
	if err != nil {
		t.Fatalf("NewSendEmailTask returned error: %v", err)
	}
	if task.Type() != TypeSendEmail {
		t.Fatalf("expected type %q, got %q", TypeSendEmail, task.Type())
	}

	var out SendEmailPayload
	if err := json.Unmarshal(task.Payload(), &out); err != nil {
		t.Fatalf("payload did not round-trip: %v", err)
	}
	if out != in {
		t.Fatalf("payload mismatch: got %+v, want %+v", out, in)
	}
}
