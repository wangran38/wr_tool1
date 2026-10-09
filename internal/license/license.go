package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Seed must be replaced with a private value before shipping: anyone who has the
// binary and this seed can mint activation codes offline.
const Seed = "CHANGE-ME-OFFLINE-ACTIVATION-SEED"

const licenseFileName = "license.json"

// The activation record lives under %APPDATA% so a Program Files install stays
// writable for users without admin rights.
const configDirName = "WR Tool"

type record struct {
	MachineCode string    `json:"machineCode"`
	LicenseKey  string    `json:"licenseKey"`
	ActivatedAt time.Time `json:"activatedAt"`
}

// Sign is the whole activation algorithm: Issue mints the code for a machine,
// Verify recomputes it. Both rely on Seed.
func Sign(machineCode string) string {
	mac := hmac.New(sha256.New, []byte(Seed))
	mac.Write([]byte(ungroup(machineCode)))
	sum := mac.Sum(nil)
	return group(strings.TrimRight(base32.StdEncoding.EncodeToString(sum[:]), "=")[:25], 5)
}

// Issue runs on the vendor side to produce the activation code for a machine code.
func Issue(machineCode string) string {
	return Sign(machineCode)
}

func Verify(machineCode, licenseKey string) bool {
	expected := Sign(machineCode)
	return subtle.ConstantTimeCompare([]byte(ungroup(expected)), []byte(ungroup(licenseKey))) == 1
}

func StorePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configDirName, licenseFileName), nil
}

func Activate(machineCode, licenseKey string) error {
	if !Verify(machineCode, licenseKey) {
		return fmt.Errorf("invalid activation code for this machine")
	}

	path, err := StorePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record{
		MachineCode: machineCode,
		LicenseKey:  ungroup(licenseKey),
		ActivatedAt: time.Now(),
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func Status() (bool, error) {
	path, err := StorePath()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	var r record
	if err := json.Unmarshal(data, &r); err != nil {
		return false, nil
	}
	current, err := MachineCode()
	if err != nil {
		return false, err
	}
	return ungroup(r.MachineCode) == ungroup(current) && Verify(current, r.LicenseKey), nil
}
