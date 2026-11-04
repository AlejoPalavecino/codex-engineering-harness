package eval

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func snapshotDirectoryAlias(t *testing.T, target, alias string) {
	t.Helper()
	if err := os.Symlink(target, alias); err == nil {
		t.Cleanup(func() { _ = os.Remove(alias) })
		return
	} else if runtime.GOOS != "windows" {
		t.Fatalf("create directory symlink: %v", err)
	}
	if output, err := exec.Command("cmd", "/c", "mklink", "/J", alias, target).CombinedOutput(); err != nil {
		t.Fatalf("create directory junction: %v: %s", err, output)
	}
	t.Cleanup(func() { _ = os.Remove(alias) })
}

func TestSaveSnapshotRejectsWorkspaceThroughDirectoryAlias(t *testing.T) {
	for _, ancestor := range []bool{false, true} {
		t.Run(map[bool]string{false: "workspace-alias", true: "workspace-parent-alias"}[ancestor], func(t *testing.T) {
			root := t.TempDir()
			workspace := filepath.Join(root, "workspace")
			if err := os.Mkdir(workspace, 0700); err != nil {
				t.Fatal(err)
			}
			alias := filepath.Join(t.TempDir(), "alias")
			target, output := workspace, filepath.Join(alias, "snapshot.json")
			if ancestor {
				target = root
				output = filepath.Join(alias, "workspace", "snapshot.json")
			}
			snapshotDirectoryAlias(t, target, alias)
			if _, err := SaveSnapshot(workspace, output); err == nil {
				t.Error("accepted snapshot output inside workspace through directory alias")
			}
			if _, err := os.Stat(filepath.Join(workspace, "snapshot.json")); !os.IsNotExist(err) {
				t.Fatalf("workspace was mutated: stat error = %v", err)
			}
		})
	}
}

func TestSaveSnapshotPreservesExistingOutput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "snapshot.json")
	const original = "existing evidence\n"
	if err := os.WriteFile(output, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := SaveSnapshot(t.TempDir(), output); err == nil {
		t.Fatal("accepted existing output")
	}
	data, err := os.ReadFile(output)
	if err != nil || string(data) != original {
		t.Fatalf("existing output changed: %q, %v", data, err)
	}
}

func TestSaveSnapshotConcurrentCreationHasOneWinner(t *testing.T) {
	workspace := t.TempDir()
	output := filepath.Join(t.TempDir(), "snapshot.json")
	const writers = 32
	start := make(chan struct{})
	results := make(chan error, writers)
	var ready sync.WaitGroup
	ready.Add(writers)
	for range writers {
		go func() {
			ready.Done()
			<-start
			_, err := SaveSnapshot(workspace, output)
			results <- err
		}()
	}
	ready.Wait()
	close(start)
	winners := 0
	for range writers {
		if <-results == nil {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("expected exactly one creator, got %d", winners)
	}
	if _, err := ReadSnapshot(output); err != nil {
		t.Fatalf("winner did not write a valid snapshot: %v", err)
	}
}
