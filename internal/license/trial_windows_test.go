//go:build windows

package license

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const testAnchorKey = `Software\WR Tool AnchorTest`

func useTestAnchor(t *testing.T) {
	t.Helper()

	t.Setenv(anchorKeyEnv, testAnchorKey)
	t.Cleanup(func() {
		exec.Command("reg", "delete", `HKCU\`+testAnchorKey, "/f").Run()
	})
	exec.Command("reg", "delete", `HKCU\`+testAnchorKey, "/f").Run()
}

func TestTrialAnchorOutlivesTheConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	useTestAnchor(t)

	if days, err := TrialDaysLeft(); err != nil || days != TrialDays {
		t.Fatalf("fresh trial should report %d days, got %d (err %v)", TrialDays, days, err)
	}

	// Customer deletes %APPDATA%\WR Tool to reset the trial.
	if err := os.RemoveAll(filepath.Join(dir, "WR Tool")); err != nil {
		t.Fatal(err)
	}
	days, err := TrialDaysLeft()
	if err != nil {
		t.Fatal(err)
	}
	if days != TrialDays {
		t.Fatalf("trial must keep running from the registry anchor, got %d days", days)
	}

	// An old anchor wins over a fresh file: backdating the file cannot renew anything.
	writeAnchor(time.Now().AddDate(0, 0, -TrialDays-1))
	if days, err = TrialDaysLeft(); err != nil || days != 0 {
		t.Fatalf("old anchor should spend the trial, got %d days (err %v)", days, err)
	}
}
