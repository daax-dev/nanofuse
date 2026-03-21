---
id: task-32
title: 'T009: Define Recording Events and Types'
status: Done
assignee:
  - '@backend-engineer'
created_date: '2025-12-22 23:16'
updated_date: '2025-12-23 00:50'
labels:
  - phase-2
  - recording
  - events
  - protobuf
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Define event types and serialization for session recording.

**Context**: Part of Phase 2 - foundation for recording system.
**No dependencies**: Can be developed in parallel with T008.

**Files to Create**:
- `internal/recording/events.go`
- `internal/recording/events_test.go`
- `api/recording.proto` (protobuf definitions)

**Event Types**:
- SESSION_START: VM boot, session metadata
- SESSION_END: Graceful shutdown
- TERMINAL_INPUT: User keystrokes
- TERMINAL_OUTPUT: Command output
- FILE_READ/WRITE: File operations (optional)
- NETWORK_REQUEST/RESPONSE: Network activity (optional)
- CHECKPOINT: Periodic sync point for seek

**Serialization**: Protobuf for efficiency (10x faster than JSON)

**Event Structure**:
```protobuf
message RecordingEvent {
  string vm_id = 1;
  string session_id = 2;
  uint64 timestamp_ns = 3;
  EventType type = 4;
  bytes payload = 5;
  map<string, string> metadata = 6;
}
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 RecordingEvent struct with all required fields
- [x] #2 EventType enum with all event types
- [x] #3 Protobuf definitions in api/recording.proto
- [x] #4 Serialization/deserialization functions
- [x] #5 Event validation functions
- [x] #6 Unit tests for serialization roundtrip
- [x] #7 Benchmark: serialization under 1ms for typical event
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented recording events package:

**Files created:**
- `internal/recording/events.go` - Core event types and serialization
- `internal/recording/events_test.go` - Comprehensive test suite
- `api/proto/recording.proto` - Protobuf definitions
- `api/proto/recording.pb.go` - Generated protobuf code

**Key implementation details:**
- EventType enum with 10 types (UNSPECIFIED + 9 event types from ADR-002)
- RecordingEvent struct with VMID, SessionID, TimestampNs, Type, Payload, Metadata
- Binary framing with 4-byte length prefix for stream processing
- Protobuf-based serialization using google.golang.org/protobuf
- Full validation functions for VMID, SessionID, EventType
- ToProto/FromProto conversion functions

**Performance:**
- MarshalBinary: ~653ns (target: <1ms) ✅
- UnmarshalBinary: ~674ns (target: <1ms) ✅
- Roundtrip: ~1.4µs ✅

**Code quality:**
- go vet passes with no issues
- All 10 tests passing
- 4 benchmarks included
<!-- SECTION:NOTES:END -->
