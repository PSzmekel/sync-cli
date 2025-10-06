package dir

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

// helper: sorted copy
func sorted(ss []string) []string {
	cp := make([]string, len(ss))
	copy(cp, ss)
	sort.Strings(cp)
	return cp
}

func TestCompareDirs_Shallow_NewUpdatedDeleted(t *testing.T) {
	src := t.TempDir()
	tgt := t.TempDir()

	// top-level files
	write := func(p string, data string, mode os.FileMode) {
		if err := os.WriteFile(p, []byte(data), mode); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	// source: a.txt (new), b.txt (same as target), sub/x.txt (ignored in shallow)
	write(filepath.Join(src, "a.txt"), "hello", 0o644)
	write(filepath.Join(src, "b.txt"), "same", 0o644)
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil { //nolint: gosec
		t.Fatalf("mkdir sub: %v", err)
	}
	write(filepath.Join(src, "sub", "x.txt"), "nested", 0o644)

	// target: b.txt (same), c.txt (deleted), sub/y.txt (ignored in shallow)
	write(filepath.Join(tgt, "b.txt"), "same", 0o644)
	write(filepath.Join(tgt, "c.txt"), "only-target", 0o644)
	if err := os.MkdirAll(filepath.Join(tgt, "sub"), 0o755); err != nil { //nolint: gosec
		t.Fatalf("mkdir sub tgt: %v", err)
	}
	write(filepath.Join(tgt, "sub", "y.txt"), "nested-target", 0o644)

	// Ustaw identyczny mtime dla b.txt po obu stronach, aby nie był Updated.
	mt := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(filepath.Join(src, "b.txt"), mt, mt); err != nil {
		t.Fatalf("chtimes src b: %v", err)
	}
	if err := os.Chtimes(filepath.Join(tgt, "b.txt"), mt, mt); err != nil {
		t.Fatalf("chtimes tgt b: %v", err)
	}

	diff, errs := CompareDirs(src, tgt, true, false)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}

	wantNew := []string{"a.txt"}
	wantUpd := []string{}        // b.txt same content
	wantDel := []string{"c.txt"} // sub/y.txt ignored in shallow

	if !reflect.DeepEqual(sorted(diff.New), sorted(wantNew)) {
		t.Fatalf("New got=%v want=%v", diff.New, wantNew)
	}
	if !reflect.DeepEqual(sorted(diff.Updated), sorted(wantUpd)) {
		t.Fatalf("Updated got=%v want=%v", diff.Updated, wantUpd)
	}
	if !reflect.DeepEqual(sorted(diff.Deleted), sorted(wantDel)) {
		t.Fatalf("Deleted got=%v want=%v", diff.Deleted, wantDel)
	}
}

func TestCompareDirs_UpdatedBySizeAndMtime(t *testing.T) {
	src := t.TempDir()
	tgt := t.TempDir()

	// size difference
	if err := os.WriteFile(filepath.Join(src, "u.txt"), []byte("1234567890"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tgt, "u.txt"), []byte("12345"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write: %v", err)
	}

	// same size but newer mtime in source
	if err := os.WriteFile(filepath.Join(src, "t.txt"), []byte("equal"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tgt, "t.txt"), []byte("equal"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write: %v", err)
	}
	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(filepath.Join(tgt, "t.txt"), older, older); err != nil {
		t.Fatalf("chtimes tgt: %v", err)
	}
	if err := os.Chtimes(filepath.Join(src, "t.txt"), newer, newer); err != nil {
		t.Fatalf("chtimes src: %v", err)
	}

	diff, errs := CompareDirs(src, tgt, false, false)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}

	wantUpd := []string{"t.txt", "u.txt"}
	if !reflect.DeepEqual(sorted(diff.Updated), sorted(wantUpd)) {
		t.Fatalf("Updated got=%v want=%v", diff.Updated, wantUpd)
	}
	if len(diff.New) != 0 || len(diff.Deleted) != 0 {
		t.Fatalf("expected only Updated; got New=%v Deleted=%v", diff.New, diff.Deleted)
	}
}

func TestCompareDirs_Deep_NewAndDeletedNested(t *testing.T) {
	src := t.TempDir()
	tgt := t.TempDir()

	// source nested new file
	if err := os.MkdirAll(filepath.Join(src, "d1", "d2"), 0o755); err != nil { //nolint: gosec
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "d1", "d2", "n.txt"), []byte("n"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write: %v", err)
	}

	// target nested deleted file
	if err := os.MkdirAll(filepath.Join(tgt, "d1"), 0o755); err != nil { //nolint: gosec
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tgt, "d1", "old.txt"), []byte("old"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write: %v", err)
	}

	diff, errs := CompareDirs(src, tgt, true, true)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}

	relNew := filepath.Join("d1", "d2", "n.txt")
	relDel := filepath.Join("d1", "old.txt")

	if !reflect.DeepEqual(sorted(diff.New), []string{relNew}) {
		t.Fatalf("New got=%v want=%v", diff.New, []string{relNew})
	}
	if !reflect.DeepEqual(sorted(diff.Deleted), []string{relDel}) {
		t.Fatalf("Deleted got=%v want=%v", diff.Deleted, []string{relDel})
	}
	if len(diff.Updated) != 0 {
		t.Fatalf("expected no Updated; got %v", diff.Updated)
	}
}

func TestCompareDirs_DeleteMissingFalse_IgnoresDeleted(t *testing.T) {
	src := t.TempDir()
	tgt := t.TempDir()

	if err := os.WriteFile(filepath.Join(tgt, "only_target.txt"), []byte("x"), 0o644); err != nil { //nolint: gosec
		t.Fatalf("write: %v", err)
	}

	diff, errs := CompareDirs(src, tgt, false, false)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs: %v", errs)
	}
	if len(diff.Deleted) != 0 {
		t.Fatalf("expected no Deleted when deleteMissing=false; got %v", diff.Deleted)
	}
}

func TestCompareDirs_MissingDirs_ReturnErrors(t *testing.T) {
	src := filepath.Join(t.TempDir(), "does_not_exist_src")
	tgt := filepath.Join(t.TempDir(), "does_not_exist_tgt")

	// deep: no source or target should return errors
	_, errs := CompareDirs(src, tgt, true, true)
	if len(errs) == 0 {
		t.Fatalf("expected errors for missing source/target in deep mode")
	}

	// shallow: no source or target should return errors
	_, errs2 := CompareDirs(src, tgt, true, false)
	if len(errs2) == 0 {
		t.Fatalf("expected errors for missing source/target in shallow mode")
	}
}
