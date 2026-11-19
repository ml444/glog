package handler

import (
	"os"
	"path/filepath"
	"sync"
)

// Logger replacement can briefly have two file handlers for the same target.
// Coordinate their I/O so a new handler never recovers an in-progress rotation
// and an older handler never acknowledges an append through a renamed file FD.
type fileCoordinator struct {
	mu   sync.Mutex
	refs int
	last *FileHandler
}

var fileCoordinators = struct {
	sync.Mutex
	paths map[string]*fileCoordinator
}{paths: make(map[string]*fileCoordinator)}

func retainFileCoordinator(path string) (string, *fileCoordinator) {
	key, err := filepath.Abs(path)
	if err != nil {
		key = filepath.Clean(path)
	}
	// Resolve existing ancestors too, so /tmp/new and /private/tmp/new retain
	// the same owner before and after the new directory has been created.
	ancestor := filepath.Dir(key)
	var missing []string
	for {
		if resolved, err := filepath.EvalSymlinks(ancestor); err == nil {
			for index := len(missing) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, missing[index])
			}
			key = filepath.Join(resolved, filepath.Base(key))
			break
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			break
		}
		missing = append(missing, filepath.Base(ancestor))
		ancestor = parent
	}
	fileCoordinators.Lock()
	defer fileCoordinators.Unlock()
	coordinator := fileCoordinators.paths[key]
	if coordinator == nil {
		coordinator = &fileCoordinator{}
		fileCoordinators.paths[key] = coordinator
	}
	coordinator.refs++
	return key, coordinator
}

func releaseFileCoordinator(key string, coordinator *fileCoordinator) {
	fileCoordinators.Lock()
	defer fileCoordinators.Unlock()
	coordinator.refs--
	if coordinator.refs == 0 {
		delete(fileCoordinators.paths, key)
	}
}

func refreshRotator(rotator IRotator) error {
	switch r := rotator.(type) {
	case *SizeRotator:
		return refreshRotatorFile(r.fs, &r.file, r.filePath, &r.curSize)
	case *TimeRotator:
		return refreshRotatorFile(r.fs, &r.file, r.filePath, nil)
	case *TimeAndSizeRotator:
		return refreshRotatorFile(r.fs, &r.file, r.filePath, &r.curSize)
	}
	return nil
}

func refreshRotatorFile(fs *rotationFS, file **os.File, path string, size *int64) error {
	if fs.err != nil || *file == nil {
		return fs.err
	}
	opened, err := (*file).Stat()
	if err != nil {
		return fs.fail(path, "stat", err)
	}
	current, err := os.Stat(path)
	if os.IsNotExist(err) || err == nil && !os.SameFile(opened, current) {
		if err := fs.closeFile(file); err != nil {
			return err
		}
		if size != nil {
			*size = 0
		}
		return nil
	}
	if err != nil {
		return fs.fail(path, "stat", err)
	}
	if size != nil {
		*size = current.Size()
	}
	return nil
}
