// Package types exports public types for lifecycle.
package types

import (
	"context"
	"time"
)

// StartupHook is a hook that is called when service is started.
// This hook is called with specified startup context. Error chanel is provided
// to report health status of service. If error is sent to error channel
// service could be restarted depends on configuration.
type StartupHook func(context.Context, chan<- error) error

// ShutdownHook is a hook that is called when service is stopped.
// This hook is called with specified shutdown context.
type ShutdownHook func(context.Context) error

// ServiceStatus represents current status of service.
//
//go:generate stringer -type=ServiceStatus -trimprefix=ServiceStatus
type ServiceStatus int

// All service statuses values.
const (
	ServiceStatusInit ServiceStatus = iota
	ServiceStatusStarting
	ServiceStatusRunning
	ServiceStatusStopping
	ServiceStatusStopped
	ServiceStatusError
)

// ServiceRestartPolicy represents rules for service restart on runtime errors.
type ServiceRestartPolicy struct {
	// RestartOnFailure indicates that service should be restarted on failure.
	RestartOnFailure bool
	// RestartDelay is a delay between restart attempts.
	RestartDelay time.Duration
	// RestartCount is a maximum number of restart attempts.
	RestartCount int
}

// DefaultRestartPolicy is a default restart policy for services.
var DefaultRestartPolicy = ServiceRestartPolicy{
	RestartOnFailure: true,
	RestartCount:     4,
	RestartDelay:     time.Millisecond * 100,
}

// ServiceConfig represents lifecycle service configuration.
type ServiceConfig struct {
	// StartupHook is a hook that is called when service is started.
	StartupHook StartupHook
	// ShutdownHook is a hook that is called when service is stopped.
	ShutdownHook ShutdownHook

	// Name of the service.
	Name string
	// RestartPolicy is a restart policy for service.
	RestartPolicy ServiceRestartPolicy
}
