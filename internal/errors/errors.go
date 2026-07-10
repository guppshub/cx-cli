package errors

import "errors"

// Standard library delegations to keep the package drop-in compatible.
var (
	New = errors.New
	Is  = errors.Is
	As  = errors.As
)

// Custom CX sentinel errors.
var (
	ErrWorkspaceNotFound  = errors.New("workspace not found")
	ErrDuplicateWorkspace = errors.New("workspace already exists")
	ErrWorkspaceActive    = errors.New("workspace is active")
)
