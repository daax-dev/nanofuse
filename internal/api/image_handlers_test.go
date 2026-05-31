package api

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/daax-dev/nanofuse/internal/logging"
	"github.com/daax-dev/nanofuse/internal/storage"
	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

func TestExecutePullJobWithRuntimeImageProviderClosesProgressGoroutine(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "nanofuse.db"))
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close: %v", err)
		}
	}()

	logger, err := logging.New(logging.Config{Level: "error"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}

	jobID := "job-runtime-provider"
	job := &types.ImagePullJob{
		ID:        jobID,
		ImageRef:  "docker.io/library/alpine:3.20",
		State:     types.PullJobPending,
		CreatedAt: time.Now(),
	}
	if err := db.CreatePullJob(job); err != nil {
		t.Fatalf("CreatePullJob: %v", err)
	}

	server := &Server{
		db:     db,
		logger: logger,
		runtimeManager: &runtimeImageProviderStub{
			image: &types.Image{
				Digest:       "sha256:test",
				Tags:         []string{"docker.io/library/alpine:3.20"},
				Architecture: "arm64",
				PulledAt:     time.Now(),
			},
		},
	}

	server.executePullJob(jobID, job.ImageRef)

	got, err := db.GetPullJob(jobID)
	if err != nil {
		t.Fatalf("GetPullJob: %v", err)
	}
	if got.State != types.PullJobCompleted {
		t.Fatalf("job state = %q, want %q", got.State, types.PullJobCompleted)
	}
	if got.ResultDigest == nil || *got.ResultDigest != "sha256:test" {
		t.Fatalf("result digest = %v, want sha256:test", got.ResultDigest)
	}
	if pullProgressGoroutineActive() {
		t.Fatal("runtime image-provider pull left the progress goroutine blocked")
	}
}

func pullProgressGoroutineActive() bool {
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	return strings.Contains(string(buf[:n]), "executePullJob.func1")
}

type runtimeImageProviderStub struct {
	image       *types.Image
	err         error
	killErr     error
	deleteErr   error
	execResult  *types.VMExecResult
	execErr     error
	killCalls   int
	deleteCalls int
	execCalls   int
	execCommand []string
}

func (r *runtimeImageProviderStub) ResolveImage(_ string) (*types.Image, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.image == nil {
		return nil, errors.New("missing test image")
	}
	return r.image, nil
}

func (r *runtimeImageProviderStub) ListImages() ([]*types.Image, error) {
	if r.image == nil {
		return nil, nil
	}
	return []*types.Image{r.image}, nil
}

func (r *runtimeImageProviderStub) SetProcessExitHandler(vmm.ProcessExitHandler) {}

func (r *runtimeImageProviderStub) Start(*types.VM, *types.Image) error { return nil }

func (r *runtimeImageProviderStub) Stop(*types.VM, int) error { return nil }

func (r *runtimeImageProviderStub) Kill(*types.VM) error {
	r.killCalls++
	return r.killErr
}

func (r *runtimeImageProviderStub) Delete(*types.VM) error {
	r.deleteCalls++
	return r.deleteErr
}

func (r *runtimeImageProviderStub) Pause(*types.VM) error { return nil }

func (r *runtimeImageProviderStub) Resume(*types.VM) error { return nil }

func (r *runtimeImageProviderStub) CreateSnapshot(*types.VM, string, string) error { return nil }

func (r *runtimeImageProviderStub) GetConsoleLogs(*types.VM, int) ([]byte, error) { return nil, nil }

func (r *runtimeImageProviderStub) CleanupNetwork(*types.VM) error { return nil }

func (r *runtimeImageProviderStub) Exec(_ context.Context, _ *types.VM, command []string) (*types.VMExecResult, error) {
	r.execCalls++
	r.execCommand = append([]string(nil), command...)
	if r.execErr != nil {
		return nil, r.execErr
	}
	if r.execResult != nil {
		return r.execResult, nil
	}
	return &types.VMExecResult{Command: command}, nil
}
