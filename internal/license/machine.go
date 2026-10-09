package license

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// MachineID is a stable per-installation identifier used as the activation seed.
// It is built from the CPU, the baseboard and the system UUID, so replacing the
// motherboard changes the machine code and a new activation code is required.
func MachineID() (string, error) {
	parts := hardwareIDs()
	if len(parts) == 0 {
		return "", fmt.Errorf("cannot derive a machine fingerprint")
	}
	return strings.Join(parts, "|"), nil
}

// MachineCode is the 16-character code shown to the customer. The WMI lookup is
// memoized because the UI re-asks it on every status refresh and every compare.
func MachineCode() (string, error) {
	machineOnce.Do(func() {
		id, err := MachineID()
		if err != nil {
			cachedErr = err
			return
		}
		sum := sha256.Sum256([]byte(id))
		code := strings.TrimRight(base32.StdEncoding.EncodeToString(sum[:]), "=")
		cachedCode = group(code[:16], 4)
	})
	return cachedCode, cachedErr
}

var (
	machineOnce sync.Once
	cachedCode  string
	cachedErr   error
)

// Keys are serial values with dots, spaces, dashes and underscores removed.
var placeholderSerials = map[string]bool{
	"none": true, "tobefilledbyoem": true, "defaultstring": true,
	"systemserialnumber": true, "notspecified": true, "notavailable": true, "0": true,
}

var serialCleaner = strings.NewReplacer(".", "", " ", "", "-", "", "_", "")

func isPlaceholder(v string) bool {
	v = strings.ToLower(serialCleaner.Replace(v))
	if v == "" || placeholderSerials[v] {
		return true
	}
	// A UUID of all F or all 0 is an unfilled firmware slot.
	return len(v) == 32 && (strings.Count(v, "f") == 32 || strings.Count(v, "0") == 32)
}

func hardwareIDs() []string {
	if runtime.GOOS != "windows" {
		return fallbackIDs()
	}

	const query = `(Get-CimInstance Win32_Processor | Select-Object -First 1 -ExpandProperty ProcessorId);` +
		`(Get-CimInstance Win32_BaseBoard | Select-Object -First 1 -ExpandProperty SerialNumber);` +
		`(Get-CimInstance Win32_ComputerSystemProduct | Select-Object -First 1 -ExpandProperty UUID)`

	out, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", query).Output()
	if err != nil {
		return fallbackIDs()
	}

	var parts []string
	for _, line := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if v := strings.TrimSpace(line); !isPlaceholder(v) {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return fallbackIDs()
	}
	return parts
}

// fallbackIDs keeps activation usable where WMI is unavailable or locked down.
func fallbackIDs() []string {
	var parts []string
	if h, err := os.Hostname(); err == nil {
		parts = append(parts, h)
	}
	if home, err := os.UserHomeDir(); err == nil {
		parts = append(parts, home)
	}
	return parts
}

func group(s string, size int) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && i%size == 0 {
			b.WriteByte('-')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func ungroup(s string) string {
	return strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(s))
}
