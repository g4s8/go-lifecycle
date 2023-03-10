// Code generated by "stringer -type=ServiceStatus -trimprefix=ServiceStatus"; DO NOT EDIT.

package types

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ServiceStatusInit-0]
	_ = x[ServiceStatusStarting-1]
	_ = x[ServiceStatusRunning-2]
	_ = x[ServiceStatusStopping-3]
	_ = x[ServiceStatusStopped-4]
	_ = x[ServiceStatusError-5]
}

const _ServiceStatus_name = "InitStartingRunningStoppingStoppedError"

var _ServiceStatus_index = [...]uint8{0, 4, 12, 19, 27, 34, 39}

func (i ServiceStatus) String() string {
	if i < 0 || i >= ServiceStatus(len(_ServiceStatus_index)-1) {
		return "ServiceStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ServiceStatus_name[_ServiceStatus_index[i]:_ServiceStatus_index[i+1]]
}
