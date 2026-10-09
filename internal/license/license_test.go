package license

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActivateStoresUnderAppData(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())

	code, err := MachineCode()
	if err != nil {
		t.Fatalf("machine code: %v", err)
	}
	if err := Activate(code, Issue(code)); err != nil {
		t.Fatalf("activate: %v", err)
	}

	path, err := StorePath()
	if err != nil {
		t.Fatalf("store path: %v", err)
	}
	if filepath.Base(path) != licenseFileName {
		t.Fatalf("unexpected file name %q", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("record not written: %v", err)
	}

	activated, err := Status()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !activated {
		t.Fatal("status should report activated")
	}

	if err := Activate(code, "WRONG-WRONG-WRONG-WRONG-WRONG"); err == nil {
		t.Fatal("tampered key accepted")
	}
}
