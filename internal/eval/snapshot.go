package eval

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Snapshot struct {
	Schema   int               `json:"schema"`
	Files    map[string]string `json:"files"`
	TreeHash string            `json:"tree_sha256"`
}

type Change struct {
	Path   string
	Status string // ADDED | REMOVED | MODIFIED
}

func digestTree(files map[string]string) string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	sum := sha256.New()
	for _, path := range paths {
		_, _ = io.WriteString(sum, path+"\x00"+files[path]+"\x00")
	}
	return hex.EncodeToString(sum.Sum(nil))
}

// SnapshotWorkspace is strictly read-only and ignores .git internals. It
// includes new/untracked files and excludes no source files automatically.
func SnapshotWorkspace(root string) (Snapshot, error) {
	if root == "" {
		return Snapshot{}, errors.New("workspace missing")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return Snapshot{}, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return Snapshot{}, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Snapshot{}, errors.New("workspace must be a real directory, not a symlink")
	}
	files := make(map[string]string)
	err = filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink rejected in snapshot: %s", d.Name())
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("unsupported workspace file: %s", d.Name())
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		h := sha256.New()
		_, copyErr := io.Copy(h, f)
		closeErr := f.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		files[filepath.ToSlash(rel)] = hex.EncodeToString(h.Sum(nil))
		return nil
	})
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Schema: Schema, Files: files, TreeHash: digestTree(files)}, nil
}

func SaveSnapshot(workspace, out string) (Snapshot, error) {
	if out == "" {
		return Snapshot{}, errors.New("destination missing")
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return Snapshot{}, err
	}
	absWorkspace, err = filepath.EvalSymlinks(absWorkspace)
	if err != nil {
		return Snapshot{}, err
	}
	absOut, err := filepath.Abs(out)
	if err != nil {
		return Snapshot{}, err
	}
	// Resolve the existing parent, including symlinks and Windows junctions.
	// The output itself need not exist and will be created exclusively below.
	parent, err := filepath.EvalSymlinks(filepath.Dir(absOut))
	if err != nil {
		return Snapshot{}, err
	}
	// Some Windows reparse points remain unresolved by EvalSymlinks and
	// appear as irregular files. Reject them instead of following an alias.
	parentInfo, err := os.Lstat(parent)
	if err != nil {
		return Snapshot{}, err
	}
	if !parentInfo.IsDir() || parentInfo.Mode()&os.ModeSymlink != 0 {
		return Snapshot{}, errors.New("snapshot output parent must resolve to a real directory")
	}
	absOut = filepath.Join(parent, filepath.Base(absOut))
	// Writing a snapshot inside the workspace would change the very evidence it records.
	rel, err := filepath.Rel(absWorkspace, absOut)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return Snapshot{}, errors.New("snapshot output must be outside the measured workspace")
	}
	s, err := SnapshotWorkspace(workspace)
	if err != nil {
		return Snapshot{}, err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return Snapshot{}, err
	}
	// Do not create directories automatically: the user chooses an existing safe output path.
	f, err := os.OpenFile(absOut, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return Snapshot{}, fmt.Errorf("snapshot already exists (refusing overwrite): %w", err)
		}
		return Snapshot{}, err
	}
	_, writeErr := f.Write(append(data, '\n'))
	closeErr := f.Close()
	if writeErr != nil {
		return Snapshot{}, writeErr
	}
	if closeErr != nil {
		return Snapshot{}, closeErr
	}
	return s, nil
}

func ReadSnapshot(path string) (Snapshot, error) {
	var snap Snapshot
	if err := decodeJSON(path, &snap); err != nil {
		return Snapshot{}, err
	}
	if snap.Schema != Schema || snap.Files == nil || len(snap.TreeHash) != 64 {
		return Snapshot{}, errors.New("invalid snapshot schema, file map or hash")
	}
	for name, value := range snap.Files {
		if !validPath(name) || len(value) != 64 {
			return Snapshot{}, fmt.Errorf("invalid snapshot entry %q", name)
		}
		if _, err := hex.DecodeString(value); err != nil {
			return Snapshot{}, fmt.Errorf("invalid SHA-256 for %q", name)
		}
	}
	if snap.TreeHash != digestTree(snap.Files) {
		return Snapshot{}, errors.New("snapshot tree SHA-256 does not match files")
	}
	return snap, nil
}

func CompareSnapshots(before, after Snapshot) []Change {
	var changes []Change
	for name, oldHash := range before.Files {
		newHash, ok := after.Files[name]
		switch {
		case !ok:
			changes = append(changes, Change{Path: name, Status: "REMOVED"})
		case oldHash != newHash:
			changes = append(changes, Change{Path: name, Status: "MODIFIED"})
		}
	}
	for name := range after.Files {
		if _, exists := before.Files[name]; !exists {
			changes = append(changes, Change{Path: name, Status: "ADDED"})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}
