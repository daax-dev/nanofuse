package vmm

import "errors"

// ErrUnsupportedOperation marks runtime features that a backend does not expose.
var ErrUnsupportedOperation = errors.New("unsupported runtime operation")
