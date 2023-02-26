// Package adaptors provides adaptors for common services.
package adaptors

import "github.com/g4s8/go-lifecycle/pkg/types"

// LifecycleRegistry provides method to register lifecycle service.
type LifecycleRegistry interface {
	RegisterService(service types.ServiceConfig)
}
