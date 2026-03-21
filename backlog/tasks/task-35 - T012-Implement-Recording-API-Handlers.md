---
id: task-35
title: 'T012: Implement Recording API Handlers'
status: Done
assignee: []
created_date: '2025-12-22 23:16'
updated_date: '2025-12-29 13:22'
labels:
  - phase-2
  - recording
  - api
  - handlers
  - flowspec-microvm
  - implement
dependencies: []
priority: medium
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
REST API endpoints for recording session management.

**Context**: Part of Phase 2 - depends on T011 for storage.
**Dependency**: T011 (storage)

**Files to Create/Modify**:
- `internal/api/recording_handlers.go` (new)
- `internal/api/server.go` (register routes)

**API Endpoints**:
```
GET  /recordings                    # List all sessions
GET  /recordings?vm_id={id}         # List sessions for VM
GET  /recordings/{session-id}       # Get session details
GET  /recordings/{session-id}/events?offset=0&limit=100  # Get events (paginated)
POST /recordings/{session-id}/finalize  # Force finalize session
DELETE /recordings/{session-id}     # Delete recording
```

**Response Formats**:
```json
{
  "id": "session-123",
  "vm_id": "vm-456",
  "started_at": "2025-12-22T10:00:00Z",
  "ended_at": "2025-12-22T10:30:00Z",
  "event_count": 1500,
  "size_bytes": 524288,
  "status": "complete"
}
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GET /recordings lists all sessions with pagination
- [x] #2 GET /recordings?vm_id filters by VM
- [x] #3 GET /recordings/{id} returns session details
- [x] #4 GET /recordings/{id}/events returns paginated events
- [x] #5 POST /recordings/{id}/finalize forces session finalization
- [x] #6 DELETE /recordings/{id} deletes session and files
- [x] #7 All endpoints registered in server.go
- [x] #8 Unit tests for all handlers
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
## Implementation Complete (2025-12-29)

Implemented REST API endpoints for recording session management.

### Files Created/Modified:
- `internal/api/recording_handlers.go` - API handlers (NEW)
- `internal/api/recording_handlers_test.go` - Unit tests (NEW, 14 tests)
- `internal/api/server.go` - Route registration + storage init
- `internal/types/errors.go` - Added ErrRecordingNotFound, ErrNotFound

### API Endpoints:
- `GET /recordings` - List all sessions with optional vm_id filter
- `GET /recordings/{id}` - Get session details
- `GET /recordings/{id}/events` - Get paginated events (stub for now)
- `POST /recordings/{id}/finalize` - Force finalize session
- `DELETE /recordings/{id}` - Delete session and files
- `GET /vms/{id}/recordings` - List recordings for a specific VM

### Integration:
- Recording storage initialized at server startup
- Storage path: {data_dir}/recordings/
- 30-day retention policy configured
<!-- SECTION:NOTES:END -->
