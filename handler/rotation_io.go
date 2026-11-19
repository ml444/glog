package handler

import (
	"os"
	"path/filepath"
)

// rotationFS stops further namespace changes after an I/O failure. Repeating a
// partially completed numbered-backup rotation could otherwise delete backups
// that are still inside the retention window.
type rotationFS struct {
	err        error
	syncFile   func(*os.File) error
	syncDir    func(string) error
	renameFile func(string, string) error
	removeFile func(string) error
}

func newRotationFS() *rotationFS {
	fs := &rotationFS{
		syncFile:   func(file *os.File) error { return file.Sync() },
		syncDir:    func(string) error { return nil },
		renameFile: os.Rename,
		removeFile: os.Remove,
	}
	if durablePlatformSupported {
		fs.syncFile = syncDurableFile
		fs.syncDir = syncDirectory
	}
	return fs
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	syncErr := syncDurableFile(dir)
	closeErr := dir.Close()
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

func (fs *rotationFS) fail(path, stage string, err error) error {
	if err != nil && fs.err == nil {
		fs.err = fileError(path, stage, err)
	}
	return fs.err
}

func (fs *rotationFS) mkdir(path string) error {
	if fs.err != nil {
		return fs.err
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return fs.fail(path, "mkdir", err)
	}
	parent := filepath.Dir(path)
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fs.fail(path, "mkdir", &os.PathError{Op: "mkdir", Path: path, Err: os.ErrExist})
		}
		return fs.fail(parent, "sync-directory", fs.syncDir(parent))
	}
	if !os.IsNotExist(err) || parent == path {
		return fs.fail(path, "mkdir", err)
	}
	if err := fs.mkdir(parent); err != nil {
		return err
	}
	if err := os.Mkdir(path, 0o755); err != nil && !os.IsExist(err) {
		return fs.fail(path, "mkdir", err)
	}
	return fs.fail(parent, "sync-directory", fs.syncDir(parent))
}

func (fs *rotationFS) open(path string) (*os.File, error) {
	if fs.err != nil {
		return nil, fs.err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return nil, fs.fail(path, "open", err)
	}
	if err := fs.syncFile(file); err != nil {
		_ = file.Close()
		return nil, fs.fail(path, "sync", err)
	}
	if err := fs.syncDir(filepath.Dir(path)); err != nil {
		_ = file.Close()
		return nil, fs.fail(path, "sync-directory", err)
	}
	return file, nil
}

func (fs *rotationFS) rename(oldPath, newPath string) error {
	if fs.err != nil {
		return fs.err
	}
	// Windows ordinary logging keeps its prior replacement behavior. Supported
	// durable platforms use atomic replacement without an unlink-before-rename.
	if !durablePlatformSupported {
		if err := fs.removeFile(newPath); err != nil && !os.IsNotExist(err) {
			return fs.fail(newPath, "remove", err)
		}
	}
	if err := fs.renameFile(oldPath, newPath); err != nil {
		return fs.fail(oldPath, "rename", err)
	}
	if err := fs.syncDir(filepath.Dir(oldPath)); err != nil {
		return fs.fail(oldPath, "sync-directory", err)
	}
	if filepath.Dir(newPath) != filepath.Dir(oldPath) {
		return fs.fail(newPath, "sync-directory", fs.syncDir(filepath.Dir(newPath)))
	}
	return nil
}

func (fs *rotationFS) remove(path string) error {
	if fs.err != nil {
		return fs.err
	}
	if err := fs.removeFile(path); err != nil {
		return fs.fail(path, "remove", err)
	}
	return fs.fail(path, "sync-directory", fs.syncDir(filepath.Dir(path)))
}

func (fs *rotationFS) closeFile(file **os.File) error {
	if *file == nil {
		return fs.err
	}
	path := (*file).Name()
	syncErr := fs.syncFile(*file)
	closeErr := (*file).Close()
	*file = nil
	if syncErr != nil {
		return fs.fail(path, "sync", syncErr)
	}
	if closeErr != nil {
		return fs.fail(path, "close", closeErr)
	}
	return fs.err
}
