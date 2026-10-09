package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wr_tool/internal/cleaner"
	"wr_tool/internal/license"
)

const testAnchorRoot = `Software\WR Tool Test`

// TestMain keeps the registry anchor used by CompareFiles out of the real
// HKCU\Software\WR Tool key, which would otherwise pin this machine's trial start.
func TestMain(m *testing.M) {
	os.Setenv("WRTOOL_TRIAL_KEY", testAnchorRoot)
	deleteAnchorKey(testAnchorRoot)
	code := m.Run()
	deleteAnchorKey(testAnchorRoot)
	os.Exit(code)
}

func deleteAnchorKey(key string) {
	exec.Command("reg", "delete", `HKCU\`+key, "/f").Run()
}

// isolateTrial gives a test its own config dir and its own anchor key, so trial
// state cannot leak between tests through the registry.
func isolateTrial(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("APPDATA", dir)

	key := testAnchorRoot + `\` + strings.NewReplacer("/", "-", " ", "-").Replace(t.Name())
	t.Setenv("WRTOOL_TRIAL_KEY", key)
	deleteAnchorKey(key)
	t.Cleanup(func() { deleteAnchorKey(key) })
	return dir
}

func setTrialStart(t *testing.T, dir string, started time.Time) {
	t.Helper()

	path := filepath.Join(dir, "WR Tool", "trial.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]time.Time{"startedAt": started})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTrialStartsOnFirstRunAndBlocksWhenOver(t *testing.T) {
	dir := isolateTrial(t)

	a := NewApp()

	st, err := a.LicenseStatus()
	if err != nil {
		t.Fatal(err)
	}
	if st.Activated || st.Expired || st.TrialDays != license.TrialDays {
		t.Fatalf("fresh install should be in trial with %d days, got %+v", license.TrialDays, st)
	}

	oldPath := filepath.Join(dir, "old.docx")
	newPath := filepath.Join(dir, "new.docx")
	writeDocx(t, oldPath, `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>甲</w:t></w:r></w:p></w:body></w:document>`)
	writeDocx(t, newPath, `<?xml version="1.0"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>乙</w:t></w:r></w:p></w:body></w:document>`)
	if _, err := a.CompareFiles(oldPath, newPath, cleaner.Options{}); err != nil {
		t.Fatalf("compare during trial should succeed: %v", err)
	}

	setTrialStart(t, dir, time.Now().AddDate(0, 0, -license.TrialDays-1))
	st, err = a.LicenseStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !st.Expired || st.TrialDays != 0 {
		t.Fatalf("past-due trial should be expired, got %+v", st)
	}

	if _, err := a.CompareFiles(oldPath, newPath, cleaner.Options{}); !errors.Is(err, ErrTrialExpired) {
		t.Fatalf("expected ErrTrialExpired, got %v", err)
	}

	// An activation lifts the block again.
	code, err := license.MachineCode()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.VerifyLicense(license.Issue(code)); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CompareFiles(oldPath, newPath, cleaner.Options{}); err != nil {
		t.Fatalf("compare after activation should succeed: %v", err)
	}
}

func TestTrialUnparseableRecordCountsAsSpent(t *testing.T) {
	dir := isolateTrial(t)

	path := filepath.Join(dir, "WR Tool", "trial.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	// A hand-edited file usually ends up looking like this: UTF-8 BOM plus junk.
	if err := os.WriteFile(path, append([]byte{0xEF, 0xBB, 0xBF}, []byte("startedAt=whatever")...), 0o600); err != nil {
		t.Fatal(err)
	}

	days, err := license.TrialDaysLeft()
	if err != nil {
		t.Fatal(err)
	}
	if days != 0 {
		t.Fatalf("unparseable record must not restart the clock, got %d days", days)
	}
}
