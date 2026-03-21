package storage

const schema = `
CREATE TABLE IF NOT EXISTS vms (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE,
	state TEXT NOT NULL,
	image_ref TEXT NOT NULL,
	image_digest TEXT NOT NULL,
	architecture TEXT NOT NULL,
	config_json TEXT NOT NULL,
	runtime_json TEXT,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	locked_by TEXT,
	locked_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vms_state ON vms(state);
CREATE INDEX IF NOT EXISTS idx_vms_name ON vms(name);
CREATE INDEX IF NOT EXISTS idx_vms_locked_by ON vms(locked_by);

CREATE TABLE IF NOT EXISTS snapshots (
	id TEXT PRIMARY KEY,
	vm_id TEXT NOT NULL,
	name TEXT,
	memory_file_path TEXT NOT NULL,
	snapshot_file_path TEXT NOT NULL,
	size_bytes INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	FOREIGN KEY (vm_id) REFERENCES vms(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_snapshots_vm_id ON snapshots(vm_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_created_at ON snapshots(created_at);

CREATE TABLE IF NOT EXISTS images (
	digest TEXT PRIMARY KEY,
	tags_json TEXT NOT NULL,
	architecture TEXT NOT NULL,
	size_bytes INTEGER NOT NULL,
	kernel_version TEXT,
	rootfs_path TEXT NOT NULL,
	kernel_path TEXT NOT NULL,
	pulled_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_images_pulled_at ON images(pulled_at);

CREATE TABLE IF NOT EXISTS image_pull_jobs (
	id TEXT PRIMARY KEY,
	image_ref TEXT NOT NULL,
	state TEXT NOT NULL,
	progress_json TEXT,
	error TEXT,
	result_digest TEXT,
	created_at TIMESTAMP NOT NULL,
	completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_image_pull_jobs_state ON image_pull_jobs(state);
CREATE INDEX IF NOT EXISTS idx_image_pull_jobs_created_at ON image_pull_jobs(created_at);

CREATE TABLE IF NOT EXISTS layer_cache (
	digest TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	version TEXT,
	type TEXT NOT NULL,
	source_url TEXT NOT NULL,
	local_path TEXT NOT NULL,
	size_bytes INTEGER NOT NULL,
	fetched_at TIMESTAMP NOT NULL,
	last_used_at TIMESTAMP NOT NULL,
	metadata_json TEXT
);

CREATE INDEX IF NOT EXISTS idx_layer_cache_name ON layer_cache(name);
CREATE INDEX IF NOT EXISTS idx_layer_cache_last_used ON layer_cache(last_used_at);

CREATE TABLE IF NOT EXISTS recording_sessions (
	id TEXT PRIMARY KEY,
	vm_id TEXT NOT NULL,
	started_at TIMESTAMP NOT NULL,
	ended_at TIMESTAMP,
	event_count INTEGER NOT NULL DEFAULT 0,
	size_bytes INTEGER NOT NULL DEFAULT 0,
	storage_path TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active',
	compressed INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_recording_sessions_vm_id ON recording_sessions(vm_id);
CREATE INDEX IF NOT EXISTS idx_recording_sessions_status ON recording_sessions(status);
CREATE INDEX IF NOT EXISTS idx_recording_sessions_started_at ON recording_sessions(started_at);
`
