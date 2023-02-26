package lifecycle

import (
	"strconv"
	"strings"

	"github.com/g4s8/go-lifecycle/pkg/types"
)

// ServiceState is a named state of service with unique ID of this server and optional error.
type ServiceState struct {
	// Service ID, is assigned on register calls.
	ID int
	// Service name, specified by user.
	Name string
	// Service status.
	Status types.ServiceStatus
	// Service error, if any.
	Error error
}

func (s ServiceState) String() string {
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(s.ID))
	sb.WriteString(" ")
	sb.WriteString(s.Name)
	sb.WriteString(": ")
	sb.WriteString(s.Status.String())
	if s.Error != nil {
		sb.WriteString(" (")
		sb.WriteString(s.Error.Error())
		sb.WriteString(")")
	}
	return sb.String()
}
