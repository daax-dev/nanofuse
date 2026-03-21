---
id: task-34
title: 'T011: Implement Recording Storage'
status: Done
assignee: []
created_date: '2025-12-22 23:16'
updated_date: '2025-12-29 13:05'
labels:
  - phase-2
  - recording
  - storage
  - local
  - flowspec-microvm
  - implement
dependencies: []
priority: high
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Storage backends for recording events with local filesystem implementation.

**Context**: Part of Phase 2 - depends on T009 for event types.
**Dependency**: T009 (events)

**Files to Create**:
- `internal/recording/storage.go` (interface)
- `internal/recording/local.go` (local backend)
- `internal/recording/storage_test.go`
- `internal/storage/schema.go` (extend with recording_sessions table)

**Storage Interface**:
```go
type RecordingStorage interface {
  StartSession(vmID, sessionID string) error
  Write(ctx context.Context, events []RecordingEvent) error
  Finalize(ctx context.Context, sessionID string) error
  GetSession(sessionID string) (*RecordingSession, error)
  ListSessions(vmID string) ([]RecordingSession, error)
  DeleteSession(sessionID string) error
}
```

**Local Storage**:
- Path: /var/lib/nanofuse/recordings/{session-id}/
- Format: events.pb (protobuf stream), metadata.json
- Compression: zstd for completed sessions

**Database Schema**:
```sql
CREATE TABLE recording_sessions (
  id TEXT PRIMARY KEY,
  vm_id TEXT NOT NULL,
  started_at TIMESTAMP,
  ended_at TIMESTAMP,
  event_count INTEGER,
  size_bytes INTEGER,
  storage_path TEXT,
  status TEXT
);
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 RecordingStorage interface defined
- [x] #2 LocalStorage writes to /var/lib/nanofuse/recordings/
- [x] #3 Database schema for recording_sessions table
- [x] #4 StartSession creates session directory and DB record
- [x] #5 Write appends events to session file
- [x] #6 Finalize compresses session and updates status
- [x] #7 Configurable retention policy (days)
- [x] #8 Unit tests for all storage operations
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation Complete (2025-12-29)

Implemented local filesystem storage backend for recording sessions.

### Files Created:
- `internal/recording/local.go` - LocalStorage implementation
- `internal/recording/local_test.go` - Storage unit tests (11 tests)
- `internal/storage/schema.go` - Added recording_sessions table

### Key Features:
- Session directory structure: /var/lib/nanofuse/recordings/{session-id}/
- Events stored in events.bin (binary format)
- Zstd compression on finalization
- Configurable retention policy with automatic cleanup
- Session metadata in metadata.json

### Storage Interface:
- StartSession() - creates session directory and files
- Write() - appends events to session (auto-creates if needed)
- Finalize() - compresses session and updates status
- GetSession() / ListSessions() - query session metadata
- DeleteSession() - remove session and files
- CleanupExpired() - retention policy enforcement
<!-- SECTION:NOTES:END -->
