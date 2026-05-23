package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daax-dev/nanofuse/internal/types"
	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// runMigrations applies database migrations
func runMigrations(conn *sql.DB) error {
	// Migration 1: Add architecture column to vms table if it doesn't exist
	var columnExists bool
	err := conn.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('vms')
		WHERE name = 'architecture'
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("failed to check for architecture column: %w", err)
	}

	if !columnExists {
		// Add architecture column with default value
		if _, err := conn.Exec(`ALTER TABLE vms ADD COLUMN architecture TEXT NOT NULL DEFAULT 'x86_64'`); err != nil {
			return fmt.Errorf("failed to add architecture column: %w", err)
		}
	}

	return nil
}

// New creates a new database connection
func New(path string) (*DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Initialize schema
	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Run migrations
	if err := runMigrations(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// CreateVM creates a new VM
func (db *DB) CreateVM(vm *types.VM) error {
	configJSON, err := json.Marshal(vm.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var runtimeJSON []byte
	if vm.Runtime != nil {
		runtimeJSON, err = json.Marshal(vm.Runtime)
		if err != nil {
			return fmt.Errorf("failed to marshal runtime: %w", err)
		}
	}

	_, err = db.conn.Exec(`
		INSERT INTO vms (id, name, state, image_ref, image_digest, architecture, config_json, runtime_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, vm.ID, vm.Name, vm.State, vm.Image, vm.ImageDigest, vm.Architecture, configJSON, runtimeJSON, vm.CreatedAt, vm.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	return nil
}

// GetVM gets a VM by ID or name
func (db *DB) GetVM(idOrName string) (*types.VM, error) {
	var vm types.VM
	var configJSON, runtimeJSON []byte
	var lockedBy sql.NullString
	var lockedAt sql.NullTime

	err := db.conn.QueryRow(`
		SELECT id, name, state, image_ref, image_digest, architecture, config_json, runtime_json,
		       created_at, updated_at, locked_by, locked_at
		FROM vms
		WHERE id = ? OR name = ?
	`, idOrName, idOrName).Scan(
		&vm.ID, &vm.Name, &vm.State, &vm.Image, &vm.ImageDigest, &vm.Architecture,
		&configJSON, &runtimeJSON, &vm.CreatedAt, &vm.UpdatedAt,
		&lockedBy, &lockedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	if err := json.Unmarshal(configJSON, &vm.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if runtimeJSON != nil {
		var runtime types.VMRuntime
		if err := json.Unmarshal(runtimeJSON, &runtime); err != nil {
			return nil, fmt.Errorf("failed to unmarshal runtime: %w", err)
		}
		vm.Runtime = &runtime
	}

	if lockedBy.Valid {
		vm.LockedBy = &lockedBy.String
	}
	if lockedAt.Valid {
		vm.LockedAt = &lockedAt.Time
	}

	return &vm, nil
}

// ListVMs lists all VMs, optionally filtered by state
func (db *DB) ListVMs(stateFilter string) ([]*types.VM, error) {
	query := `
		SELECT id, name, state, image_ref, image_digest, architecture, config_json, runtime_json,
		       created_at, updated_at, locked_by, locked_at
		FROM vms
	`
	args := []interface{}{}

	if stateFilter != "" {
		query += " WHERE state = ?"
		args = append(args, stateFilter)
	}

	query += " ORDER BY created_at DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}
	defer rows.Close()

	var vms []*types.VM
	for rows.Next() {
		var vm types.VM
		var configJSON, runtimeJSON []byte
		var lockedBy sql.NullString
		var lockedAt sql.NullTime

		err := rows.Scan(
			&vm.ID, &vm.Name, &vm.State, &vm.Image, &vm.ImageDigest, &vm.Architecture,
			&configJSON, &runtimeJSON, &vm.CreatedAt, &vm.UpdatedAt,
			&lockedBy, &lockedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan VM: %w", err)
		}

		if err := json.Unmarshal(configJSON, &vm.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		if runtimeJSON != nil {
			var runtime types.VMRuntime
			if err := json.Unmarshal(runtimeJSON, &runtime); err != nil {
				return nil, fmt.Errorf("failed to unmarshal runtime: %w", err)
			}
			vm.Runtime = &runtime
		}

		if lockedBy.Valid {
			vm.LockedBy = &lockedBy.String
		}
		if lockedAt.Valid {
			vm.LockedAt = &lockedAt.Time
		}

		vms = append(vms, &vm)
	}

	return vms, nil
}

// UpdateVM updates a VM
func (db *DB) UpdateVM(vm *types.VM) error {
	configJSON, err := json.Marshal(vm.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var runtimeJSON []byte
	if vm.Runtime != nil {
		runtimeJSON, err = json.Marshal(vm.Runtime)
		if err != nil {
			return fmt.Errorf("failed to marshal runtime: %w", err)
		}
	}

	vm.UpdatedAt = time.Now()

	_, err = db.conn.Exec(`
		UPDATE vms
		SET state = ?, config_json = ?, runtime_json = ?, updated_at = ?, locked_by = ?, locked_at = ?
		WHERE id = ?
	`, vm.State, configJSON, runtimeJSON, vm.UpdatedAt, vm.LockedBy, vm.LockedAt, vm.ID)

	if err != nil {
		return fmt.Errorf("failed to update VM: %w", err)
	}

	return nil
}

// DeleteVM deletes a VM
func (db *DB) DeleteVM(id string) error {
	_, err := db.conn.Exec("DELETE FROM vms WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}
	return nil
}

// AcquireLock attempts to acquire a lock on a VM
func (db *DB) AcquireLock(vmID, operation string) error {
	result, err := db.conn.Exec(`
		UPDATE vms
		SET locked_by = ?, locked_at = ?
		WHERE id = ? AND (locked_by IS NULL OR locked_at < ?)
	`, operation, time.Now(), vmID, time.Now().Add(-5*time.Minute))

	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check lock result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("VM is locked by another operation")
	}

	return nil
}

// ReleaseLock releases a lock on a VM
func (db *DB) ReleaseLock(vmID string) error {
	_, err := db.conn.Exec(`
		UPDATE vms
		SET locked_by = NULL, locked_at = NULL
		WHERE id = ?
	`, vmID)

	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

// CreateSnapshot creates a new snapshot
func (db *DB) CreateSnapshot(snapshot *types.Snapshot) error {
	_, err := db.conn.Exec(`
		INSERT INTO snapshots (id, vm_id, name, memory_file_path, snapshot_file_path, size_bytes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, snapshot.ID, snapshot.VMID, snapshot.Name, snapshot.MemoryFilePath,
		snapshot.SnapshotFilePath, snapshot.SizeBytes, snapshot.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	return nil
}

// GetSnapshot gets a snapshot by ID
func (db *DB) GetSnapshot(id string) (*types.Snapshot, error) {
	var snapshot types.Snapshot
	var name sql.NullString

	err := db.conn.QueryRow(`
		SELECT id, vm_id, name, memory_file_path, snapshot_file_path, size_bytes, created_at
		FROM snapshots
		WHERE id = ?
	`, id).Scan(&snapshot.ID, &snapshot.VMID, &name, &snapshot.MemoryFilePath,
		&snapshot.SnapshotFilePath, &snapshot.SizeBytes, &snapshot.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	if name.Valid {
		snapshot.Name = name.String
	}

	return &snapshot, nil
}

// ListSnapshots lists all snapshots for a VM
func (db *DB) ListSnapshots(vmID string) ([]*types.Snapshot, error) {
	rows, err := db.conn.Query(`
		SELECT id, vm_id, name, memory_file_path, snapshot_file_path, size_bytes, created_at
		FROM snapshots
		WHERE vm_id = ?
		ORDER BY created_at DESC
	`, vmID)

	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*types.Snapshot
	for rows.Next() {
		var snapshot types.Snapshot
		var name sql.NullString

		err := rows.Scan(&snapshot.ID, &snapshot.VMID, &name, &snapshot.MemoryFilePath,
			&snapshot.SnapshotFilePath, &snapshot.SizeBytes, &snapshot.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}

		if name.Valid {
			snapshot.Name = name.String
		}

		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, nil
}

// DeleteSnapshot deletes a snapshot
func (db *DB) DeleteSnapshot(id string) error {
	_, err := db.conn.Exec("DELETE FROM snapshots WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}
	return nil
}

// CreateImage creates a new image
func (db *DB) CreateImage(image *types.Image) error {
	tagsJSON, err := json.Marshal(image.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = db.conn.Exec(`
		INSERT INTO images (digest, tags_json, architecture, size_bytes, kernel_version, rootfs_path, kernel_path, pulled_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, image.Digest, tagsJSON, image.Architecture, image.SizeBytes,
		image.KernelVersion, image.RootfsPath, image.KernelPath, image.PulledAt)

	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}

	return nil
}

// UpsertImage creates or updates an image (idempotent)
func (db *DB) UpsertImage(image *types.Image) error {
	tagsJSON, err := json.Marshal(image.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	_, err = db.conn.Exec(`
		INSERT INTO images (digest, tags_json, architecture, size_bytes, kernel_version, rootfs_path, kernel_path, pulled_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(digest) DO UPDATE SET
			tags_json = excluded.tags_json,
			architecture = excluded.architecture,
			size_bytes = excluded.size_bytes,
			kernel_version = excluded.kernel_version,
			rootfs_path = excluded.rootfs_path,
			kernel_path = excluded.kernel_path,
			pulled_at = excluded.pulled_at
	`, image.Digest, tagsJSON, image.Architecture, image.SizeBytes,
		image.KernelVersion, image.RootfsPath, image.KernelPath, image.PulledAt)

	if err != nil {
		return fmt.Errorf("failed to upsert image: %w", err)
	}

	return nil
}

// GetImage gets an image by digest
func (db *DB) GetImage(digest string) (*types.Image, error) {
	var image types.Image
	var tagsJSON []byte
	var kernelVersion sql.NullString

	err := db.conn.QueryRow(`
		SELECT digest, tags_json, architecture, size_bytes, kernel_version, rootfs_path, kernel_path, pulled_at
		FROM images
		WHERE digest = ?
	`, digest).Scan(&image.Digest, &tagsJSON, &image.Architecture, &image.SizeBytes,
		&kernelVersion, &image.RootfsPath, &image.KernelPath, &image.PulledAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	if err := json.Unmarshal(tagsJSON, &image.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if kernelVersion.Valid {
		image.KernelVersion = kernelVersion.String
	}

	return &image, nil
}

// GetImageByTag gets an image by tag reference
func (db *DB) GetImageByTag(tag string) (*types.Image, error) {
	// Query all images and search their tags
	// SQLite JSON functions can be tricky, so we'll search in Go
	rows, err := db.conn.Query(`
		SELECT digest, tags_json, architecture, size_bytes, kernel_version, rootfs_path, kernel_path, pulled_at
		FROM images
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var image types.Image
		var tagsJSON []byte
		var kernelVersion sql.NullString

		if err := rows.Scan(&image.Digest, &tagsJSON, &image.Architecture, &image.SizeBytes,
			&kernelVersion, &image.RootfsPath, &image.KernelPath, &image.PulledAt); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &image.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if kernelVersion.Valid {
			image.KernelVersion = kernelVersion.String
		}

		// Check if this image has the requested tag
		for _, imageTag := range image.Tags {
			if imageTag == tag {
				return &image, nil
			}
		}
	}

	// No image found with this tag
	return nil, nil
}

// ListImages lists all images
func (db *DB) ListImages() ([]*types.Image, error) {
	rows, err := db.conn.Query(`
		SELECT digest, tags_json, architecture, size_bytes, kernel_version, rootfs_path, kernel_path, pulled_at
		FROM images
		ORDER BY pulled_at DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	defer rows.Close()

	var images []*types.Image
	for rows.Next() {
		var image types.Image
		var tagsJSON []byte
		var kernelVersion sql.NullString

		err := rows.Scan(&image.Digest, &tagsJSON, &image.Architecture, &image.SizeBytes,
			&kernelVersion, &image.RootfsPath, &image.KernelPath, &image.PulledAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}

		if err := json.Unmarshal(tagsJSON, &image.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if kernelVersion.Valid {
			image.KernelVersion = kernelVersion.String
		}

		images = append(images, &image)
	}

	return images, nil
}

// DeleteImage deletes an image
func (db *DB) DeleteImage(digest string) error {
	_, err := db.conn.Exec("DELETE FROM images WHERE digest = ?", digest)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// CreatePullJob creates a new pull job
func (db *DB) CreatePullJob(job *types.ImagePullJob) error {
	var progressJSON []byte
	var err error
	if job.Progress != nil {
		progressJSON, err = json.Marshal(job.Progress)
		if err != nil {
			return fmt.Errorf("failed to marshal progress: %w", err)
		}
	}

	_, err = db.conn.Exec(`
		INSERT INTO image_pull_jobs (id, image_ref, state, progress_json, error, result_digest, created_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, job.ID, job.ImageRef, job.State, progressJSON, job.Error, job.ResultDigest, job.CreatedAt, job.CompletedAt)

	if err != nil {
		return fmt.Errorf("failed to create pull job: %w", err)
	}

	return nil
}

// GetPullJob gets a pull job by ID
func (db *DB) GetPullJob(id string) (*types.ImagePullJob, error) {
	var job types.ImagePullJob
	var progressJSON []byte
	var errorStr sql.NullString
	var resultDigest sql.NullString
	var completedAt sql.NullTime

	err := db.conn.QueryRow(`
		SELECT id, image_ref, state, progress_json, error, result_digest, created_at, completed_at
		FROM image_pull_jobs
		WHERE id = ?
	`, id).Scan(&job.ID, &job.ImageRef, &job.State, &progressJSON, &errorStr,
		&resultDigest, &job.CreatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pull job: %w", err)
	}

	if progressJSON != nil {
		var progress types.PullProgress
		if err := json.Unmarshal(progressJSON, &progress); err != nil {
			return nil, fmt.Errorf("failed to unmarshal progress: %w", err)
		}
		job.Progress = &progress
	}

	if errorStr.Valid {
		job.Error = &errorStr.String
	}
	if resultDigest.Valid {
		job.ResultDigest = &resultDigest.String
	}
	if completedAt.Valid {
		job.CompletedAt = &completedAt.Time
	}

	return &job, nil
}

// UpdatePullJob updates a pull job
func (db *DB) UpdatePullJob(job *types.ImagePullJob) error {
	var progressJSON []byte
	var err error
	if job.Progress != nil {
		progressJSON, err = json.Marshal(job.Progress)
		if err != nil {
			return fmt.Errorf("failed to marshal progress: %w", err)
		}
	}

	_, err = db.conn.Exec(`
		UPDATE image_pull_jobs
		SET state = ?, progress_json = ?, error = ?, result_digest = ?, completed_at = ?
		WHERE id = ?
	`, job.State, progressJSON, job.Error, job.ResultDigest, job.CompletedAt, job.ID)

	if err != nil {
		return fmt.Errorf("failed to update pull job: %w", err)
	}

	return nil
}

// PutCachedLayer stores or updates a cached layer
func (db *DB) PutCachedLayer(digest, name, version, layerType, sourceURL, localPath string, sizeBytes int64, fetchedAt, lastUsedAt time.Time, metadataJSON string) error {
	_, err := db.conn.Exec(`
		INSERT INTO layer_cache (digest, name, version, type, source_url, local_path, size_bytes, fetched_at, last_used_at, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(digest) DO UPDATE SET
			name = excluded.name,
			version = excluded.version,
			type = excluded.type,
			source_url = excluded.source_url,
			local_path = excluded.local_path,
			size_bytes = excluded.size_bytes,
			fetched_at = excluded.fetched_at,
			last_used_at = excluded.last_used_at,
			metadata_json = excluded.metadata_json
	`, digest, name, version, layerType, sourceURL, localPath, sizeBytes, fetchedAt, lastUsedAt, metadataJSON)

	if err != nil {
		return fmt.Errorf("failed to put cached layer: %w", err)
	}

	return nil
}

// GetCachedLayer retrieves a cached layer by digest
func (db *DB) GetCachedLayer(digest string) (name, version, layerType, sourceURL, localPath string, sizeBytes int64, fetchedAt, lastUsedAt time.Time, metadataJSON string, found bool, err error) {
	var versionNull sql.NullString
	var metadataNull sql.NullString

	err = db.conn.QueryRow(`
		SELECT name, version, type, source_url, local_path, size_bytes, fetched_at, last_used_at, metadata_json
		FROM layer_cache
		WHERE digest = ?
	`, digest).Scan(&name, &versionNull, &layerType, &sourceURL, &localPath, &sizeBytes, &fetchedAt, &lastUsedAt, &metadataNull)

	if err == sql.ErrNoRows {
		return "", "", "", "", "", 0, time.Time{}, time.Time{}, "", false, nil
	}
	if err != nil {
		return "", "", "", "", "", 0, time.Time{}, time.Time{}, "", false, fmt.Errorf("failed to get cached layer: %w", err)
	}

	if versionNull.Valid {
		version = versionNull.String
	}
	if metadataNull.Valid {
		metadataJSON = metadataNull.String
	}

	return name, version, layerType, sourceURL, localPath, sizeBytes, fetchedAt, lastUsedAt, metadataJSON, true, nil
}

// CachedLayerExists checks if a layer with the given digest exists in cache
func (db *DB) CachedLayerExists(digest string) (bool, error) {
	var count int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM layer_cache WHERE digest = ?
	`, digest).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("failed to check cached layer existence: %w", err)
	}

	return count > 0, nil
}

// TouchCachedLayer updates the last_used_at timestamp for a layer
func (db *DB) TouchCachedLayer(digest string, lastUsedAt time.Time) error {
	_, err := db.conn.Exec(`
		UPDATE layer_cache
		SET last_used_at = ?
		WHERE digest = ?
	`, lastUsedAt, digest)

	if err != nil {
		return fmt.Errorf("failed to touch cached layer: %w", err)
	}

	return nil
}

// GetCachedLayerStats returns cache statistics
func (db *DB) GetCachedLayerStats() (totalLayers int, totalBytes int64, oldestAccess time.Time, err error) {
	var oldestAccessStr sql.NullString

	err = db.conn.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(size_bytes), 0),
			MIN(last_used_at)
		FROM layer_cache
	`).Scan(&totalLayers, &totalBytes, &oldestAccessStr)

	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("failed to get cache stats: %w", err)
	}

	if oldestAccessStr.Valid {
		// Parse the timestamp string
		oldestAccess, err = time.Parse("2006-01-02 15:04:05.999999999-07:00", oldestAccessStr.String)
		if err != nil {
			// Try alternate format
			oldestAccess, err = time.Parse("2006-01-02T15:04:05Z", oldestAccessStr.String)
			if err != nil {
				// Fall back to now if parse fails
				oldestAccess = time.Now()
			}
		}
	} else {
		oldestAccess = time.Now()
	}

	return totalLayers, totalBytes, oldestAccess, nil
}

// GetOldestCachedLayers returns layers ordered by last_used_at (oldest first)
func (db *DB) GetOldestCachedLayers(limit int) ([]struct {
	Digest    string
	LocalPath string
	SizeBytes int64
}, error) {
	var rows *sql.Rows
	var err error

	if limit > 0 {
		// Use parameterized query to prevent SQL injection
		rows, err = db.conn.Query(`
			SELECT digest, local_path, size_bytes
			FROM layer_cache
			ORDER BY last_used_at ASC
			LIMIT ?
		`, limit)
	} else {
		rows, err = db.conn.Query(`
			SELECT digest, local_path, size_bytes
			FROM layer_cache
			ORDER BY last_used_at ASC
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oldest cached layers: %w", err)
	}
	defer rows.Close()

	var layers []struct {
		Digest    string
		LocalPath string
		SizeBytes int64
	}

	for rows.Next() {
		var l struct {
			Digest    string
			LocalPath string
			SizeBytes int64
		}
		if err := rows.Scan(&l.Digest, &l.LocalPath, &l.SizeBytes); err != nil {
			return nil, fmt.Errorf("failed to scan cached layer: %w", err)
		}
		layers = append(layers, l)
	}

	return layers, nil
}

// DeleteCachedLayer removes a layer from the cache
func (db *DB) DeleteCachedLayer(digest string) error {
	_, err := db.conn.Exec("DELETE FROM layer_cache WHERE digest = ?", digest)
	if err != nil {
		return fmt.Errorf("failed to delete cached layer: %w", err)
	}
	return nil
}
