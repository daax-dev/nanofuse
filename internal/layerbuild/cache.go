package layerbuild

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/daax-dev/nanofuse/internal/storage"
)

// DefaultCacheDir is the default location for the layer cache
const DefaultCacheDir = "/var/lib/nanofuse/layer-cache"

// DefaultCacheSizeLimit is the default cache size limit (10GB)
const DefaultCacheSizeLimit = 10 * 1024 * 1024 * 1024

// layerCache implements the LayerCache interface using SQLite and filesystem storage.
type layerCache struct {
	cacheDir  string
	db        *storage.DB
	sizeLimit int64
	mu        sync.RWMutex
}

// NewLayerCache creates a new layer cache instance.
//
// Parameters:
//   - cacheDir: Directory to store cached layer tarballs
//   - dbPath: Path to SQLite database for metadata
//   - sizeLimit: Maximum cache size in bytes (triggers eviction when exceeded)
//
// The cache directory will be created if it doesn't exist.
func NewLayerCache(cacheDir, dbPath string, sizeLimit int64) (LayerCache, error) {
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Open or create database
	db, err := storage.New(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open cache database: %w", err)
	}

	return &layerCache{
		cacheDir:  cacheDir,
		db:        db,
		sizeLimit: sizeLimit,
	}, nil
}

// Get retrieves a layer by its SHA256 digest.
// Returns nil, nil if the layer is not cached.
func (c *layerCache) Get(digest string) (*CachedLayer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	name, version, layerType, sourceURL, localPath, sizeBytes, fetchedAt, lastUsedAt, metadataJSON, found, err := c.db.GetCachedLayer(digest)
	if err != nil {
		return nil, fmt.Errorf("failed to get cached layer: %w", err)
	}

	if !found {
		return nil, nil
	}

	// Parse layer type
	lt := LayerType(layerType)
	if !lt.Valid() {
		return nil, fmt.Errorf("invalid layer type in cache: %s", layerType)
	}

	// Parse metadata if present
	var metadata *LayerPackage
	if metadataJSON != "" {
		metadata = &LayerPackage{}
		if err := json.Unmarshal([]byte(metadataJSON), metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Auto-touch on read for LRU tracking
	// Fire-and-forget with short timeout - cache touch is best-effort
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if err := c.db.TouchCachedLayer(digest, time.Now()); err != nil {
			// Log but don't fail - touch is non-critical for cache operation
			// The layer was already retrieved successfully
			_ = err // Acknowledged: touch failure is non-fatal
		}
	}()

	return &CachedLayer{
		Digest:     digest,
		Name:       name,
		Version:    version,
		Type:       lt,
		SourceURL:  sourceURL,
		LocalPath:  localPath,
		SizeBytes:  sizeBytes,
		FetchedAt:  fetchedAt,
		LastUsedAt: lastUsedAt,
		Metadata:   metadata,
	}, nil
}

// Put stores a layer in the cache.
// If a layer with the same digest already exists, it will be updated.
func (c *layerCache) Put(layer *CachedLayer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Validate layer type
	if !layer.Type.Valid() {
		return fmt.Errorf("invalid layer type: %s", layer.Type)
	}

	// Serialize metadata
	var metadataJSON string
	if layer.Metadata != nil {
		data, err := json.Marshal(layer.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(data)
	}

	// Copy tarball to cache directory
	cachedPath := filepath.Join(c.cacheDir, layer.Digest+".tar.gz")

	// Only copy if source is different from destination
	if layer.LocalPath != cachedPath {
		src, err := os.Open(layer.LocalPath)
		if err != nil {
			return fmt.Errorf("failed to open source tarball: %w", err)
		}
		defer src.Close()

		dst, err := os.Create(cachedPath)
		if err != nil {
			return fmt.Errorf("failed to create cached tarball: %w", err)
		}
		defer dst.Close()

		copied, err := io.Copy(dst, src)
		if err != nil {
			os.Remove(cachedPath) // Clean up partial file
			return fmt.Errorf("failed to copy tarball to cache: %w", err)
		}

		// Update size if copying revealed actual size
		if layer.SizeBytes == 0 {
			layer.SizeBytes = copied
		}
	}

	// Store metadata in database
	if err := c.db.PutCachedLayer(
		layer.Digest,
		layer.Name,
		layer.Version,
		layer.Type.String(),
		layer.SourceURL,
		cachedPath,
		layer.SizeBytes,
		layer.FetchedAt,
		layer.LastUsedAt,
		metadataJSON,
	); err != nil {
		os.Remove(cachedPath) // Clean up if database insert fails
		return fmt.Errorf("failed to store layer metadata: %w", err)
	}

	return nil
}

// Exists checks if a layer with the given digest is cached.
func (c *layerCache) Exists(digest string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	exists, err := c.db.CachedLayerExists(digest)
	if err != nil {
		return false, fmt.Errorf("failed to check layer existence: %w", err)
	}

	return exists, nil
}

// Touch updates the LastUsedAt timestamp for LRU tracking.
// This is called automatically by Get(), but can be called explicitly.
func (c *layerCache) Touch(digest string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.db.TouchCachedLayer(digest, time.Now()); err != nil {
		return fmt.Errorf("failed to touch layer: %w", err)
	}

	return nil
}

// Evict removes layers to free space using LRU policy.
// Returns the number of bytes freed.
//
// Eviction continues until at least targetBytes are freed or no more layers remain.
// Layers are removed in LRU order (oldest access time first).
func (c *layerCache) Evict(targetBytes int64) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var totalFreed int64

	// Get layers ordered by least recently used
	layers, err := c.db.GetOldestCachedLayers(0) // 0 = no limit
	if err != nil {
		return 0, fmt.Errorf("failed to get layers for eviction: %w", err)
	}

	// Evict layers until target is met
	for _, layer := range layers {
		if totalFreed >= targetBytes {
			break
		}

		// Delete from database
		if err := c.db.DeleteCachedLayer(layer.Digest); err != nil {
			return totalFreed, fmt.Errorf("failed to delete layer from database: %w", err)
		}

		// Delete tarball file
		if err := os.Remove(layer.LocalPath); err != nil && !os.IsNotExist(err) {
			// Log but continue - database is already updated
			// In production, this would use a proper logger
			fmt.Fprintf(os.Stderr, "warning: failed to delete cached file %s: %v\n", layer.LocalPath, err)
		}

		totalFreed += layer.SizeBytes
	}

	return totalFreed, nil
}

// Stats returns cache statistics.
func (c *layerCache) Stats() (*CacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalLayers, totalBytes, oldestAccess, err := c.db.GetCachedLayerStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache stats: %w", err)
	}

	return &CacheStats{
		TotalLayers:  totalLayers,
		TotalBytes:   totalBytes,
		OldestAccess: oldestAccess,
	}, nil
}

// Close closes the cache and releases resources.
func (c *layerCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.db.Close(); err != nil {
		return fmt.Errorf("failed to close cache database: %w", err)
	}

	return nil
}

// CheckAndEvict checks if the cache exceeds the size limit and evicts if necessary.
// This is a convenience method that combines Stats() and Evict().
func (c *layerCache) CheckAndEvict() (int64, error) {
	stats, err := c.Stats()
	if err != nil {
		return 0, fmt.Errorf("failed to get cache stats: %w", err)
	}

	if stats.TotalBytes > c.sizeLimit {
		excess := stats.TotalBytes - c.sizeLimit
		return c.Evict(excess)
	}

	return 0, nil
}
