package license

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"time"
)

const TrialDays = 30

const trialFileName = "trial.json"

// anchorKeyEnv lets tests and QA point the registry anchor at a throwaway key so
// running the suite does not pin the developer's own trial start.
const anchorKeyEnv = "WRTOOL_TRIAL_KEY"

const defaultAnchorKey = `Software\WR Tool`
const anchorValueName = "trialStart"

type trialRecord struct {
	StartedAt time.Time `json:"startedAt"`
}

type anchorState int

const (
	anchorAbsent anchorState = iota
	anchorSet
	anchorBroken
)

func trialPath() (string, error) {
	path, err := StorePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(path), trialFileName), nil
}

func anchorKeyPath() string {
	if v := os.Getenv(anchorKeyEnv); v != "" {
		return v
	}
	return defaultAnchorKey
}

// TrialDaysLeft starts the countdown on the first run and rounds a partial day up,
// so the customer sees the full window on day one instead of 29.
func TrialDaysLeft() (int, error) {
	start, err := trialStart()
	if err != nil {
		return 0, err
	}

	used := time.Since(start)
	if start.IsZero() || used < 0 {
		// Zero start means no anchor could be trusted; negative means the clock was
		// rolled back. Both count as spent so neither can be used to renew the trial.
		return 0, nil
	}
	remaining := time.Duration(TrialDays)*24*time.Hour - used
	if remaining <= 0 {
		return 0, nil
	}
	return int(math.Ceil(remaining.Hours() / 24)), nil
}

// trialStart reconciles the two anchors (config file + registry) and returns the
// earliest one. An anchor that exists but cannot be parsed is treated as
// tampering: without another usable value the trial counts as spent.
func trialStart() (time.Time, error) {
	path, err := trialPath()
	if err != nil {
		return time.Time{}, err
	}

	fileStart, fileState := readFileStart(path)
	regStart, regState := readAnchor()

	var starts []time.Time
	if fileState == anchorSet {
		starts = append(starts, fileStart)
	}
	if regState == anchorSet {
		starts = append(starts, regStart)
	}

	if len(starts) == 0 {
		if fileState == anchorAbsent && regState == anchorAbsent {
			return startTrial(path)
		}
		return time.Time{}, nil
	}

	earliest := starts[0]
	for _, s := range starts[1:] {
		if s.Before(earliest) {
			earliest = s
		}
	}
	if fileState != anchorSet {
		_ = writeTrialFile(path, earliest)
	}
	if regState != anchorSet {
		writeAnchor(earliest)
	}
	return earliest, nil
}

func readFileStart(path string) (time.Time, anchorState) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return time.Time{}, anchorAbsent
	}
	if err != nil {
		return time.Time{}, anchorBroken
	}

	var r trialRecord
	if json.Unmarshal(data, &r) != nil || r.StartedAt.IsZero() {
		return time.Time{}, anchorBroken
	}
	return r.StartedAt, anchorSet
}

func startTrial(path string) (time.Time, error) {
	now := time.Now()
	if err := writeTrialFile(path, now); err != nil {
		return time.Time{}, err
	}
	writeAnchor(now)
	return now, nil
}

func writeTrialFile(path string, at time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(trialRecord{StartedAt: at})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
