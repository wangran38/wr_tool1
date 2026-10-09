//go:build !windows

package license

import "time"

// The trial has no second anchor off Windows; the config file alone decides.
func readAnchor() (time.Time, anchorState) { return time.Time{}, anchorAbsent }

func writeAnchor(time.Time) {}
