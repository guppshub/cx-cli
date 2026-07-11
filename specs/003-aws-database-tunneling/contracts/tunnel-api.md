# API Contract: AWS Database Tunneling

This document defines the contract interfaces for the database connection and tunneling subsystems.

## 1. Network Boundary Dialer Contract

```go
package workflow

import (
	"context"
	"net"
)

// TunnelDialer establishes network tunnels to cloud targets.
type TunnelDialer interface {
	DialTunnel(ctx context.Context, target *TunnelTarget) (net.Conn, error)
}
```

## 2. Interactive Prompter Callback

```go
package provider

// Prompter prompts the user for authentication input (e.g. MFA, passwords).
type Prompter func(prompt string, secret bool) (string, error)
```

## 3. Concrete AWS Provider Dialer

```go
package aws

import (
	"context"
	"net"

	"github.com/guppshub/cx-cli/internal/workflow"
)

// Provider implements cloud credential verification and network tunneling.
type Provider struct {
	profile string
	region  string
}

// New creates a new AWS Provider dialer.
func New(profile, region string) *Provider

// EnsureCredentials negotiates AWS session authentication.
// Invokes the Prompter callback if MFA token verification is required.
func (p *Provider) EnsureCredentials(ctx context.Context, prompt func(string, bool) (string, error)) error

// DialTunnel launches session-manager-plugin in the background and wraps it.
func (p *Provider) DialTunnel(ctx context.Context, target *workflow.TunnelTarget) (net.Conn, error)
```

## 4. Workflow DB Forwarder

```go
package db

import (
	"context"

	"github.com/guppshub/cx-cli/internal/workflow"
)

// Controller manages the local port listener loop and state updates.
type Controller struct {
	dialer workflow.TunnelDialer
}

// NewController creates a new database tunnel controller.
func NewController(dialer workflow.TunnelDialer) *Controller

// Start binds to a local port and forwards traffic to the dialer stream.
func (c *Controller) Start(ctx context.Context, name string, localPort int, target *workflow.TunnelTarget) error
```
