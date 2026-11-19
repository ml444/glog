package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// This journal contains namespace progress only, never log records. Each move
// is idempotent while its destination cannot become a later move's source.
type sizeRotationIntent struct {
	Version     int `json:"version"`
	Next        int `json:"next"`
	BackupCount int `json:"backup_count"`
}

func (r *SizeRotator) intentPath() string {
	return sizeRotationIntentPath(r.filePath)
}

func sizeRotationIntentPath(path string) string {
	return filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".glog-rotation")
}

func checkPendingSizeRotation(path string, kind RotatorType) error {
	if !durablePlatformSupported || kind == FileRotatorTypeSize {
		return nil
	}
	marker := sizeRotationIntentPath(path)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fileError(marker, "recover-rotation", err)
	}
	return fileError(marker, "recover-rotation", errors.New("unfinished size rotation: reopen with the size rotator before changing rotation type"))
}

func (r *SizeRotator) saveIntent(intent sizeRotationIntent) error {
	path := r.intentPath()
	data, err := json.Marshal(intent)
	if err != nil {
		return r.fs.fail(path, "rotation-progress", err)
	}
	temp := path + ".tmp"
	file, err := os.OpenFile(temp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return r.fs.fail(temp, "rotation-progress", err)
	}
	_, writeErr := writeAll(file, data, func(file *os.File, buf []byte) (int, error) { return file.Write(buf) })
	if writeErr == nil {
		writeErr = r.fs.syncFile(file)
	}
	closeErr := file.Close()
	if writeErr != nil {
		return r.fs.fail(temp, "rotation-progress", writeErr)
	}
	if closeErr != nil {
		return r.fs.fail(temp, "rotation-progress", closeErr)
	}
	return r.fs.rename(temp, path)
}

func (r *SizeRotator) recoverRotation() error {
	data, err := os.ReadFile(r.intentPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return r.fs.fail(r.intentPath(), "recover-rotation", err)
	}
	var intent sizeRotationIntent
	if err := json.Unmarshal(data, &intent); err != nil {
		return r.fs.fail(r.intentPath(), "recover-rotation", err)
	}
	if intent.Version != 1 || intent.BackupCount < 1 || intent.Next < 0 || intent.Next > intent.BackupCount {
		return r.fs.fail(r.intentPath(), "recover-rotation", errors.New("invalid size rotation progress"))
	}
	_, err = r.finishRotation(intent)
	return err
}

func (r *SizeRotator) finishRotation(intent sizeRotationIntent) (*os.File, error) {
	persisted := intent.Next
	for intent.Next > 0 {
		from := r.filePath
		if intent.Next > 1 {
			from = fmt.Sprintf("%s.%d", r.filePath, intent.Next-1)
		}
		to := fmt.Sprintf("%s.%d", r.filePath, intent.Next)
		info, err := os.Stat(from)
		if err != nil && !os.IsNotExist(err) {
			return nil, r.fs.fail(from, "rotate", err)
		}
		if err == nil {
			if info.IsDir() {
				return nil, r.fs.fail(from, "rotate", errors.New("backup source is a directory"))
			}
			// Persist the exact move before executing it. Missing sources can be
			// skipped, but must not leave a checkpoint pointing above a move:
			// replay could otherwise move its newly created destination again.
			if persisted != intent.Next {
				if err := r.saveIntent(intent); err != nil {
					return nil, err
				}
				persisted = intent.Next
			}
			if err := r.fs.rename(from, to); err != nil {
				return nil, err
			}
		}
		intent.Next--
	}
	// Checkpoint the active-file rename before creating the new active file.
	// Otherwise recovery could mistake that new file for the old source and
	// overwrite the latest backup with an empty file.
	if persisted != 0 {
		if err := r.saveIntent(intent); err != nil {
			return nil, err
		}
	}
	file, err := r.fs.open(r.filePath)
	if err != nil {
		return nil, err
	}
	r.file = file
	info, err := file.Stat()
	if err != nil {
		_ = r.fs.closeFile(&r.file)
		return nil, r.fs.fail(r.filePath, "stat", err)
	}
	r.curSize = info.Size()
	if err := r.fs.remove(r.intentPath()); err != nil {
		_ = r.fs.closeFile(&r.file)
		return nil, err
	}
	return file, nil
}
