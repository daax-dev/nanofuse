package vmm

import (
	"context"

	"github.com/daax-dev/nanofuse/internal/types"
)

// ProcessExitHandler is called when a VM runtime exits.
// exitCode is nil when the runtime cannot provide a process exit code.
type ProcessExitHandler func(vmID string, exitCode *int, err error)

// Manager controls a local microVM runtime implementation.
type Manager interface {
	SetProcessExitHandler(ProcessExitHandler)
	Start(vm *types.VM, image *types.Image) error
	Stop(vm *types.VM, timeoutSeconds int) error
	Kill(vm *types.VM) error
	Delete(vm *types.VM) error
	Pause(vm *types.VM) error
	Resume(vm *types.VM) error
	CreateSnapshot(vm *types.VM, snapshotPath, memPath string) error
	LoadSnapshot(vm *types.VM, snapshotPath, memPath string) error
	GetConsoleLogs(vm *types.VM, tailLines int) ([]byte, error)
	CleanupNetwork(vm *types.VM) error
}

// ImageProvider is implemented by runtimes that own image discovery/pull.
type ImageProvider interface {
	ResolveImage(imageRef string) (*types.Image, error)
	ListImages() ([]*types.Image, error)
}

// CommandExecutor is implemented by runtimes that can execute a command in a running VM.
type CommandExecutor interface {
	Exec(ctx context.Context, vm *types.VM, command []string) (*types.VMExecResult, error)
}
