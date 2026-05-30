package recording

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// EventType represents the type of recording event
type EventType uint8

const (
	EventSessionStart EventType = iota
	EventSessionEnd
	EventTerminalInput
	EventTerminalOutput
	EventFileRead
	EventFileWrite
	EventNetworkRequest
	EventNetworkResponse
	EventCheckpoint
)

// String returns the string representation of the event type
func (e EventType) String() string {
	switch e {
	case EventSessionStart:
		return "SESSION_START"
	case EventSessionEnd:
		return "SESSION_END"
	case EventTerminalInput:
		return "TERMINAL_INPUT"
	case EventTerminalOutput:
		return "TERMINAL_OUTPUT"
	case EventFileRead:
		return "FILE_READ"
	case EventFileWrite:
		return "FILE_WRITE"
	case EventNetworkRequest:
		return "NETWORK_REQUEST"
	case EventNetworkResponse:
		return "NETWORK_RESPONSE"
	case EventCheckpoint:
		return "CHECKPOINT"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", e)
	}
}

// Event represents a single recording event from a VM session.
// Events are streamed from the recording agent inside the VM to the host
// via virtio-vsock.
//
// Wire format (little-endian):
//
//	[4 bytes] total length (excluding this field)
//	[36 bytes] VM ID (UUID string, null-padded)
//	[36 bytes] Session ID (UUID string, null-padded)
//	[8 bytes] timestamp (nanoseconds since Unix epoch)
//	[1 byte] event type
//	[4 bytes] payload length
//	[N bytes] payload
//	[4 bytes] metadata count
//	[...] metadata entries (each: 2-byte key len, key, 2-byte val len, val)
type Event struct {
	VMID      string            `json:"vm_id"`
	SessionID string            `json:"session_id"`
	Timestamp time.Time         `json:"timestamp"`
	Type      EventType         `json:"type"`
	Payload   []byte            `json:"payload"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

const (
	// UUIDFieldSize is the fixed size for UUID string fields (36 bytes + null padding)
	UUIDFieldSize = 36
	// HeaderSize is the minimum event header size before payload
	// 4 (len) + 36 (vmid) + 36 (sessionid) + 8 (ts) + 1 (type) + 4 (payload len) = 89
	HeaderSize = 89
	// MaxPayloadSize limits the maximum payload to prevent memory exhaustion
	MaxPayloadSize = 16 * 1024 * 1024 // 16 MB
	// MaxMetadataEntries limits metadata entries per event
	MaxMetadataEntries = 100
	// MaxMetadataKeySize limits metadata key length
	MaxMetadataKeySize = 256
	// MaxMetadataValueSize limits metadata value length
	MaxMetadataValueSize = 4096
)

// ReadEvent reads a single event from the stream.
// Returns io.EOF if the stream is closed cleanly.
func ReadEvent(r io.Reader) (*Event, error) {
	// Read total length
	var totalLen uint32
	if err := binary.Read(r, binary.LittleEndian, &totalLen); err != nil {
		return nil, err
	}

	// Sanity check on length
	if totalLen < HeaderSize-4 || totalLen > MaxPayloadSize+HeaderSize {
		return nil, fmt.Errorf("invalid event length: %d", totalLen)
	}

	// Read VM ID (36 bytes, null-padded)
	vmIDBytes := make([]byte, UUIDFieldSize)
	if _, err := io.ReadFull(r, vmIDBytes); err != nil {
		return nil, fmt.Errorf("failed to read VM ID: %w", err)
	}
	vmID := trimNullPadding(vmIDBytes)

	// Read Session ID (36 bytes, null-padded)
	sessionIDBytes := make([]byte, UUIDFieldSize)
	if _, err := io.ReadFull(r, sessionIDBytes); err != nil {
		return nil, fmt.Errorf("failed to read session ID: %w", err)
	}
	sessionID := trimNullPadding(sessionIDBytes)

	// Read timestamp (nanoseconds since epoch)
	var timestampNs uint64
	if err := binary.Read(r, binary.LittleEndian, &timestampNs); err != nil {
		return nil, fmt.Errorf("failed to read timestamp: %w", err)
	}

	// Read event type
	var eventType uint8
	if err := binary.Read(r, binary.LittleEndian, &eventType); err != nil {
		return nil, fmt.Errorf("failed to read event type: %w", err)
	}

	// Read payload length
	var payloadLen uint32
	if err := binary.Read(r, binary.LittleEndian, &payloadLen); err != nil {
		return nil, fmt.Errorf("failed to read payload length: %w", err)
	}

	if payloadLen > MaxPayloadSize {
		return nil, fmt.Errorf("payload too large: %d bytes", payloadLen)
	}

	// Read payload
	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
	}

	// Read metadata count
	var metadataCount uint32
	if err := binary.Read(r, binary.LittleEndian, &metadataCount); err != nil {
		return nil, fmt.Errorf("failed to read metadata count: %w", err)
	}

	if metadataCount > MaxMetadataEntries {
		return nil, fmt.Errorf("too many metadata entries: %d", metadataCount)
	}

	// Read metadata entries
	metadata := make(map[string]string, metadataCount)
	for i := uint32(0); i < metadataCount; i++ {
		key, err := readLengthPrefixedString(r, MaxMetadataKeySize)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata key: %w", err)
		}
		value, err := readLengthPrefixedString(r, MaxMetadataValueSize)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata value: %w", err)
		}
		metadata[key] = value
	}

	return &Event{
		VMID:      vmID,
		SessionID: sessionID,
		Timestamp: time.Unix(0, int64(timestampNs)), //nolint:gosec // timestamps are valid until year 2262
		Type:      EventType(eventType),
		Payload:   payload,
		Metadata:  metadata,
	}, nil
}

// WriteEvent writes a single event to the stream.
func WriteEvent(w io.Writer, e *Event) error {
	// Calculate total length
	payloadLen := len(e.Payload)
	metadataLen := calculateMetadataSize(e.Metadata)
	// Reject oversized payloads (symmetric with ReadEvent) so the uint32
	// length conversions below cannot overflow.
	if payloadLen > MaxPayloadSize {
		return fmt.Errorf("payload too large: %d bytes (max %d)", payloadLen, MaxPayloadSize)
	}
	// Bounds check: max message size is well under uint32 max (header + payload + metadata)
	totalLen := uint32(UUIDFieldSize + UUIDFieldSize + 8 + 1 + 4 + payloadLen + 4 + metadataLen) //nolint:gosec // size bounded by max payload/metadata constants

	// Write total length
	if err := binary.Write(w, binary.LittleEndian, totalLen); err != nil {
		return fmt.Errorf("failed to write length: %w", err)
	}

	// Write VM ID (null-padded to 36 bytes)
	vmIDBytes := make([]byte, UUIDFieldSize)
	copy(vmIDBytes, e.VMID)
	if _, err := w.Write(vmIDBytes); err != nil {
		return fmt.Errorf("failed to write VM ID: %w", err)
	}

	// Write Session ID (null-padded to 36 bytes)
	sessionIDBytes := make([]byte, UUIDFieldSize)
	copy(sessionIDBytes, e.SessionID)
	if _, err := w.Write(sessionIDBytes); err != nil {
		return fmt.Errorf("failed to write session ID: %w", err)
	}

	// Write timestamp
	if err := binary.Write(w, binary.LittleEndian, uint64(e.Timestamp.UnixNano())); err != nil { //nolint:gosec // timestamps are always positive
		return fmt.Errorf("failed to write timestamp: %w", err)
	}

	// Write event type
	if err := binary.Write(w, binary.LittleEndian, uint8(e.Type)); err != nil {
		return fmt.Errorf("failed to write event type: %w", err)
	}

	// Write payload length and payload
	if err := binary.Write(w, binary.LittleEndian, uint32(payloadLen)); err != nil { //nolint:gosec // bounded by MaxPayloadSize check above
		return fmt.Errorf("failed to write payload length: %w", err)
	}
	if payloadLen > 0 {
		if _, err := w.Write(e.Payload); err != nil {
			return fmt.Errorf("failed to write payload: %w", err)
		}
	}

	// Write metadata count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(e.Metadata))); err != nil { //nolint:gosec // metadata count bounded by MaxMetadataEntries constant
		return fmt.Errorf("failed to write metadata count: %w", err)
	}

	// Write metadata entries
	for key, value := range e.Metadata {
		if err := writeLengthPrefixedString(w, key); err != nil {
			return fmt.Errorf("failed to write metadata key: %w", err)
		}
		if err := writeLengthPrefixedString(w, value); err != nil {
			return fmt.Errorf("failed to write metadata value: %w", err)
		}
	}

	return nil
}

// trimNullPadding removes null bytes from the end of a byte slice
func trimNullPadding(b []byte) string {
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != 0 {
			return string(b[:i+1])
		}
	}
	return ""
}

// readLengthPrefixedString reads a 2-byte length prefix followed by the string
func readLengthPrefixedString(r io.Reader, maxLen uint16) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	if length > maxLen {
		return "", fmt.Errorf("string too long: %d > %d", length, maxLen)
	}
	data := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(r, data); err != nil {
			return "", err
		}
	}
	return string(data), nil
}

// writeLengthPrefixedString writes a 2-byte length prefix followed by the string
func writeLengthPrefixedString(w io.Writer, s string) error {
	strLen := len(s)
	if strLen > 65535 { // Max uint16
		return fmt.Errorf("string too long: %d > 65535", strLen)
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(strLen)); err != nil { //nolint:gosec // bounds checked above
		return err
	}
	if strLen > 0 {
		if _, err := w.Write([]byte(s)); err != nil {
			return err
		}
	}
	return nil
}

// calculateMetadataSize calculates the wire size of metadata
func calculateMetadataSize(metadata map[string]string) int {
	size := 0
	for key, value := range metadata {
		size += 2 + len(key) + 2 + len(value)
	}
	return size
}
