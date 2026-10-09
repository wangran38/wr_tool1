//go:build windows

package license

import (
	"time"

	"golang.org/x/sys/windows/registry"
)

// readAnchor looks at HKCU so that deleting %APPDATA%\WR Tool alone cannot restart
// the countdown. Access trouble is reported as absent: a locked-down hive must not
// lock the customer out, while a value we cannot parse is a tamper signal.
func readAnchor() (time.Time, anchorState) {
	k, err := registry.OpenKey(registry.CURRENT_USER, anchorKeyPath(), registry.READ)
	if err != nil {
		return time.Time{}, anchorAbsent
	}
	defer k.Close()

	v, _, err := k.GetStringValue(anchorValueName)
	if err != nil {
		return time.Time{}, anchorAbsent
	}

	at, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		return time.Time{}, anchorBroken
	}
	return at, anchorSet
}

func writeAnchor(at time.Time) {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, anchorKeyPath(), registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()
	_ = k.SetStringValue(anchorValueName, at.Format(time.RFC3339Nano))
}
