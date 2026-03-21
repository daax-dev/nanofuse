---
id: task-33
title: 'T010: Implement Vsock Receiver'
status: Done
assignee: []
created_date: '2025-12-22 23:16'
updated_date: '2025-12-29 13:00'
labels:
  - phase-2
  - recording
  - vsock
  - receiver
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Host-side virtio-vsock listener for receiving recording events from VMs.

**Context**: Part of Phase 2 - depends on T009 for event types.
**Dependency**: T009 (events)

**Files to Create**:
- `internal/recording/receiver.go`
- `internal/recording/receiver_test.go`

**Receiver Architecture**:
```
VM (vsock guest) ---> Host (vsock listener) ---> Event Parser ---> Storage
```

**Key Functions**:
- StartReceiver(port uint32) error
- handleConnection(conn net.Conn, vmID string)
- parseEvents(stream io.Reader) <-chan RecordingEvent

**Connection Lifecycle**:
1. VM connects to host vsock
2. Receiver accepts connection
3. Parse event stream (length-prefixed protobuf)
4. Buffer events for batch storage
5. Finalize session on disconnect

**Concurrency**: One goroutine per VM connection
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Listen on configurable vsock port (default 52)
- [x] #2 Accept connections from multiple VMs concurrently
- [x] #3 Parse incoming protobuf event stream
- [x] #4 Buffer events before writing to storage
- [x] #5 Handle VM disconnection gracefully
- [x] #6 Finalize session on connection close
- [x] #7 Unit tests with vsock mock
- [ ] #8 Integration test with real vsock connection
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation Complete (2025-12-29)

Implemented vsock receiver for recording events from VMs.

### Files Created:
- `internal/recording/events.go` - Event types and wire format
- `internal/recording/receiver.go` - Vsock receiver implementation
- `internal/recording/events_test.go` - Event serialization tests
- `internal/recording/receiver_test.go` - Receiver unit tests

### Key Features:
- Event types matching ADR-002 protobuf spec
- Binary wire format for efficient transmission
- Configurable vsock port (default 52)
- Concurrent connection handling
- Event buffering and batch storage
- Session lifecycle tracking

### Dependencies Added:
- github.com/mdlayher/vsock - Go vsock library

Note: AC #8 (integration test with real vsock) requires VM environment and is not included in unit tests.
<!-- SECTION:NOTES:END -->
