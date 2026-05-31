package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daax-dev/nanofuse/internal/types"
	"github.com/daax-dev/nanofuse/internal/vmm"
)

func writeRuntimeUnsupportedError(w http.ResponseWriter, operation string, err error, details map[string]interface{}) bool {
	if !errors.Is(err, vmm.ErrUnsupportedOperation) {
		return false
	}
	if details == nil {
		details = make(map[string]interface{})
	}
	details["operation"] = operation
	details["runtime_error"] = err.Error()
	types.WriteError(w, http.StatusNotImplemented, types.ErrUnsupportedOperation,
		fmt.Sprintf("Runtime does not support %s", operation), details)
	return true
}
