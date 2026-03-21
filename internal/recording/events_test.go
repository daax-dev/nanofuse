package recording

import (
	"bytes"
	"testing"
	"time"
)

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventSessionStart, "SESSION_START"},
		{EventSessionEnd, "SESSION_END"},
		{EventTerminalInput, "TERMINAL_INPUT"},
		{EventTerminalOutput, "TERMINAL_OUTPUT"},
		{EventFileRead, "FILE_READ"},
		{EventFileWrite, "FILE_WRITE"},
		{EventNetworkRequest, "NETWORK_REQUEST"},
		{EventNetworkResponse, "NETWORK_RESPONSE"},
		{EventCheckpoint, "CHECKPOINT"},
		{EventType(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.expected {
				t.Errorf("EventType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEventRoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Nanosecond)

	// Note: VM ID and Session ID fields are 36 bytes max (UUID length)
	tests := []struct {
		name  string
		event *Event
	}{
		{
			name: "minimal event",
			event: &Event{
				VMID:      "12345678-1234-1234-1234-123456789012",
				SessionID: "abcdefgh-1234-1234-1234-123456789012",
				Timestamp: now,
				Type:      EventSessionStart,
				Payload:   nil,
				Metadata:  nil,
			},
		},
		{
			name: "terminal output event",
			event: &Event{
				VMID:      "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
				SessionID: "ffffffff-0000-1111-2222-333333333333",
				Timestamp: now,
				Type:      EventTerminalOutput,
				Payload:   []byte("Hello, World!\n"),
				Metadata: map[string]string{
					"stream": "stdout",
					"width":  "80",
					"height": "24",
				},
			},
		},
		{
			name: "file write event",
			event: &Event{
				VMID:      "vm-test1234",
				SessionID: "ss-test5678",
				Timestamp: now,
				Type:      EventFileWrite,
				Payload:   []byte("file contents here"),
				Metadata: map[string]string{
					"path": "/tmp/test.txt",
					"mode": "0644",
				},
			},
		},
		{
			name: "large payload",
			event: &Event{
				VMID:      "vm-large",
				SessionID: "ss-large",
				Timestamp: now,
				Type:      EventTerminalOutput,
				Payload:   bytes.Repeat([]byte("x"), 100000),
				Metadata:  nil,
			},
		},
		{
			name: "empty strings",
			event: &Event{
				VMID:      "",
				SessionID: "",
				Timestamp: now,
				Type:      EventCheckpoint,
				Payload:   []byte{},
				Metadata:  map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			var buf bytes.Buffer
			if err := WriteEvent(&buf, tt.event); err != nil {
				t.Fatalf("WriteEvent failed: %v", err)
			}

			// Deserialize
			got, err := ReadEvent(&buf)
			if err != nil {
				t.Fatalf("ReadEvent failed: %v", err)
			}

			// Compare
			if got.VMID != tt.event.VMID {
				t.Errorf("VMID = %q, want %q", got.VMID, tt.event.VMID)
			}
			if got.SessionID != tt.event.SessionID {
				t.Errorf("SessionID = %q, want %q", got.SessionID, tt.event.SessionID)
			}
			if !got.Timestamp.Equal(tt.event.Timestamp) {
				t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.event.Timestamp)
			}
			if got.Type != tt.event.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.event.Type)
			}
			if !bytes.Equal(got.Payload, tt.event.Payload) {
				t.Errorf("Payload length = %d, want %d", len(got.Payload), len(tt.event.Payload))
			}
			if len(got.Metadata) != len(tt.event.Metadata) {
				t.Errorf("Metadata length = %d, want %d", len(got.Metadata), len(tt.event.Metadata))
			}
			for k, v := range tt.event.Metadata {
				if got.Metadata[k] != v {
					t.Errorf("Metadata[%q] = %q, want %q", k, got.Metadata[k], v)
				}
			}
		})
	}
}

func TestReadEventInvalidLength(t *testing.T) {
	// Create buffer with invalid length (too large)
	var buf bytes.Buffer
	buf.Write([]byte{0xFF, 0xFF, 0xFF, 0x7F}) // 2GB, way too large

	_, err := ReadEvent(&buf)
	if err == nil {
		t.Error("Expected error for invalid length, got nil")
	}
}

func TestReadEventTruncated(t *testing.T) {
	// Create a valid event and truncate it
	event := &Event{
		VMID:      "vm-test",
		SessionID: "ss-test",
		Timestamp: time.Now(),
		Type:      EventSessionStart,
		Payload:   []byte("test payload"),
	}

	var fullBuf bytes.Buffer
	if err := WriteEvent(&fullBuf, event); err != nil {
		t.Fatalf("WriteEvent failed: %v", err)
	}

	// Truncate to half
	truncated := fullBuf.Bytes()[:fullBuf.Len()/2]
	_, err := ReadEvent(bytes.NewReader(truncated))
	if err == nil {
		t.Error("Expected error for truncated data, got nil")
	}
}

func TestTrimNullPadding(t *testing.T) {
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte("hello\x00\x00\x00"), "hello"},
		{[]byte("test"), "test"},
		{[]byte("\x00\x00\x00"), ""},
		{[]byte{}, ""},
		{[]byte("a\x00b"), "a\x00b"}, // null in middle is preserved
	}

	for _, tt := range tests {
		got := trimNullPadding(tt.input)
		if got != tt.expected {
			t.Errorf("trimNullPadding(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMultipleEventsInSequence(t *testing.T) {
	events := []*Event{
		{
			VMID:      "vm-seq",
			SessionID: "ss-seq",
			Timestamp: time.Now(),
			Type:      EventSessionStart,
			Payload:   []byte("session start"),
		},
		{
			VMID:      "vm-seq",
			SessionID: "ss-seq",
			Timestamp: time.Now().Add(time.Second),
			Type:      EventTerminalOutput,
			Payload:   []byte("output 1"),
		},
		{
			VMID:      "vm-seq",
			SessionID: "ss-seq",
			Timestamp: time.Now().Add(2 * time.Second),
			Type:      EventTerminalInput,
			Payload:   []byte("input 1"),
		},
		{
			VMID:      "vm-seq",
			SessionID: "ss-seq",
			Timestamp: time.Now().Add(3 * time.Second),
			Type:      EventSessionEnd,
			Payload:   nil,
		},
	}

	// Write all events
	var buf bytes.Buffer
	for _, e := range events {
		if err := WriteEvent(&buf, e); err != nil {
			t.Fatalf("WriteEvent failed: %v", err)
		}
	}

	// Read all events back
	for i, expected := range events {
		got, err := ReadEvent(&buf)
		if err != nil {
			t.Fatalf("ReadEvent %d failed: %v", i, err)
		}
		if got.Type != expected.Type {
			t.Errorf("Event %d: Type = %v, want %v", i, got.Type, expected.Type)
		}
		if !bytes.Equal(got.Payload, expected.Payload) {
			t.Errorf("Event %d: Payload mismatch", i)
		}
	}
}

func BenchmarkWriteEvent(b *testing.B) {
	event := &Event{
		VMID:      "vm-benchmark",
		SessionID: "ss-benchmark",
		Timestamp: time.Now(),
		Type:      EventTerminalOutput,
		Payload:   bytes.Repeat([]byte("benchmark data "), 100),
		Metadata: map[string]string{
			"stream": "stdout",
		},
	}

	var buf bytes.Buffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		WriteEvent(&buf, event)
	}
}

func BenchmarkReadEvent(b *testing.B) {
	event := &Event{
		VMID:      "vm-benchmark",
		SessionID: "ss-benchmark",
		Timestamp: time.Now(),
		Type:      EventTerminalOutput,
		Payload:   bytes.Repeat([]byte("benchmark data "), 100),
		Metadata: map[string]string{
			"stream": "stdout",
		},
	}

	var buf bytes.Buffer
	WriteEvent(&buf, event)
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReadEvent(bytes.NewReader(data))
	}
}
