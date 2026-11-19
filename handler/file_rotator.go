package handler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type IRotator interface {
	NeedRollover(msg []byte) (*os.File, bool, error)
	DoRollover() (*os.File, error)
	Close() error
	// RecordBytesWritten updates size accounting after a successful Write (avoids per-write Seek).
	RecordBytesWritten(n int)
}

func NewRotator(cfg *FileHandlerConfig) (IRotator, error) {
	if cfg == nil {
		return nil, errors.New("file rotator config is nil")
	}
	switch cfg.RotatorType {
	case FileRotatorTypeSize:
		return NewSizeRotator(cfg)
	case FileRotatorTypeTime:
		return NewTimeRotator(cfg)
	case FileRotatorTypeTimeAndSize:
		return NewTimeAndSizeRotator(cfg)
	default:
		return NewTimeAndSizeRotator(cfg)
	}
}

type SizeRotator struct {
	file     *os.File
	cfg      *FileHandlerConfig
	filePath string
	maxSize  int64
	curSize  int64
}

func NewSizeRotator(cfg *FileHandlerConfig) (*SizeRotator, error) {
	if cfg == nil {
		return nil, errors.New("size rotator config is nil")
	}
	r := SizeRotator{
		cfg:     cfg,
		maxSize: cfg.MaxFileSize,
	}
	err := r.init()
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *SizeRotator) init() error {
	err := mkdir(r.cfg.FileDir)
	if err != nil {
		return err
	}
	r.filePath = r.getFilepath()
	return nil
}

func (r *SizeRotator) getFilepath() string {
	var parties []string
	if r.cfg.FileName != "" {
		parties = append(parties, r.cfg.FileName)
	}
	parties = append(parties, r.cfg.FileSuffix)
	return filepath.Join(r.cfg.FileDir, strings.Join(parties, "."))
}

func (r *SizeRotator) NeedRollover(msg []byte) (*os.File, bool, error) {
	if r.maxSize > 0 && int64(len(msg)) > r.maxSize {
		return r.file, false, ErrFileSizeLimit
	}
	if r.file == nil {
		file, err := open(r.filePath)
		if err != nil {
			return nil, false, err
		}
		r.file = file
		st, err := file.Stat()
		if err != nil {
			return nil, false, err
		}
		r.curSize = st.Size()
	}
	if r.maxSize > 0 && r.curSize+int64(len(msg)) > r.maxSize {
		return r.file, true, nil
	}
	return r.file, false, nil
}

func (r *SizeRotator) RecordBytesWritten(n int) {
	if n <= 0 {
		return
	}
	r.curSize += int64(n)
}

func (r *SizeRotator) DoRollover() (*os.File, error) {
	if r.file != nil {
		if err := closeFile(&r.file); err != nil {
			return nil, err
		}
	}
	if r.cfg.BackupCount > 0 {
		for i := r.cfg.BackupCount; i > 0; i-- {
			sfn := fmt.Sprintf("%s.%d", r.filePath, i-1)
			dfn := fmt.Sprintf("%s.%d", r.filePath, i)
			if IsFileExist(sfn) {
				if IsFileExist(dfn) {
					err := os.Remove(dfn)
					if err != nil {
						return nil, err
					}
				}
				err := os.Rename(sfn, dfn)
				if err != nil {
					return nil, err
				}
			}
		}
		dfn := fmt.Sprintf("%s.1", r.filePath)
		if IsFileExist(dfn) {
			err := os.Remove(dfn)
			if err != nil {
				return nil, err
			}
		}
		if IsFileExist(r.filePath) {
			err := os.Rename(r.filePath, dfn)
			if err != nil {
				return nil, err
			}
		}
	}

	f, err := open(r.filePath)
	if err != nil {
		return nil, err
	}
	r.file = f
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	r.curSize = st.Size()
	return f, nil
}

func (r *SizeRotator) Close() error {
	return closeFile(&r.file)
}

type TimeRotator struct {
	cfg        *FileHandlerConfig
	file       *os.File
	filePath   string
	interval   int64
	rolloverAt int64
	reCompile  *regexp.Regexp
}

func NewTimeRotator(cfg *FileHandlerConfig) (*TimeRotator, error) {
	r := &TimeRotator{cfg: cfg}
	err := r.init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *TimeRotator) init() error {
	if r.cfg == nil {
		return errors.New("rotator config is nil")
	}
	r.interval = r.cfg.Interval
	if r.interval <= 0 {
		r.interval = 60 * 60 // default 1 hour
	}
	r.rolloverAt = getRolloverSecond(r.interval)
	var err error
	r.reCompile, err = regexp.Compile(r.cfg.ReMatch)
	if err != nil {
		return err
	}
	if err := mkdir(r.cfg.FileDir); err != nil {
		return err
	}

	r.filePath = r.getFilepath()
	return nil
}

func (r *TimeRotator) getFilepath() string {
	var parties []string
	if r.cfg.FileName != "" {
		parties = append(parties, r.cfg.FileName)
	}
	parties = append(parties, r.cfg.FileSuffix)
	return filepath.Join(r.cfg.FileDir, strings.Join(parties, "."))
}

func (r *TimeRotator) NeedRollover(_ []byte) (*os.File, bool, error) {
	if r.file == nil {
		var err error
		r.file, err = open(r.filePath)
		if err != nil {
			return r.file, false, err
		}
	}
	t := time.Now().Unix()
	if t >= r.rolloverAt {
		return r.file, true, nil
	}
	return r.file, false, nil
}

func (r *TimeRotator) RecordBytesWritten(_ int) {}

func (r *TimeRotator) DoRollover() (*os.File, error) {
	if r.file != nil {
		if err := closeFile(&r.file); err != nil {
			return nil, err
		}
	}
	curTime := time.Now()
	suffixTime := curTime.Format(r.cfg.TimeSuffixFmt)
	dfn := fmt.Sprintf("%s.%s", r.filePath, suffixTime)
	if IsFileExist(r.filePath) {
		backupPath, err := availableBackupPath(dfn)
		if err != nil {
			return nil, err
		}
		err = os.Rename(r.filePath, backupPath)
		if err != nil {
			return nil, err
		}
	}
	if backupCount := r.cfg.BackupCount; backupCount > 0 {
		dir, filename := filepath.Split(r.filePath)
		delFileList, err := collectRotatedBackupPaths(dir, filename, r.reCompile)
		if err != nil {
			return nil, err
		}
		if delLen := len(delFileList); delLen > backupCount {
			sort.Strings(delFileList)
			delFileList = delFileList[:delLen-backupCount]
			for _, filePath := range delFileList {
				if err := os.Remove(filePath); err != nil {
					return nil, err
				}
			}
		}
	}

	// next rolloverAt
	r.rolloverAt = curTime.Unix() + r.interval

	f, err := open(r.filePath)
	if err != nil {
		return nil, err
	}
	r.file = f
	return f, nil
}

func (r *TimeRotator) Close() error {
	return closeFile(&r.file)
}

type TimeAndSizeRotator struct {
	cfg                *FileHandlerConfig
	file               *os.File
	filename           string
	filePath           string
	maxSize            int64
	curSize            int64
	backupCount        int
	interval           int64
	suffixFmt          string
	rolloverAt         int64
	rolloverTimeSuffix string
	reCompile          *regexp.Regexp
}

func NewTimeAndSizeRotator(cfg *FileHandlerConfig) (*TimeAndSizeRotator, error) {
	if cfg == nil {
		return nil, errors.New("time and size rotator config is nil")
	}
	r := &TimeAndSizeRotator{
		cfg:         cfg,
		filename:    cfg.FileName,
		filePath:    cfg.FileDir,
		maxSize:     cfg.MaxFileSize,
		backupCount: cfg.BackupCount,
		interval:    cfg.Interval,
		suffixFmt:   cfg.TimeSuffixFmt,
		reCompile:   nil,
	}
	err := r.init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *TimeAndSizeRotator) init() error {
	if r.cfg == nil {
		return errors.New("config is nil")
	}

	if r.interval <= 0 {
		r.interval = 60 * 60 // default 1 hour
	}
	r.rolloverAt = getRolloverSecond(r.interval)
	var err error
	r.reCompile, err = regexp.Compile(r.cfg.ReMatch)
	if err != nil {
		return err
	}
	if err := mkdir(r.cfg.FileDir); err != nil {
		return err
	}

	r.filePath = r.getNewFilepath()
	return nil
}

func (r *TimeAndSizeRotator) NeedRollover(msg []byte) (*os.File, bool, error) {
	if r.maxSize > 0 && int64(len(msg)) > r.maxSize {
		return r.file, false, ErrFileSizeLimit
	}
	if r.file == nil {
		var err error
		r.file, err = open(r.filePath)
		if err != nil {
			return r.file, false, err
		}
		modTime, err := getFileModTime(r.file)
		if err != nil {
			return r.file, false, err
		}
		if modTime.Unix()+r.interval < r.rolloverAt {
			return r.file, true, nil
		}
		st, err := r.file.Stat()
		if err != nil {
			return r.file, false, err
		}
		r.curSize = st.Size()
	}
	t := time.Now().Unix()
	if t >= r.rolloverAt {
		return r.file, true, nil
	}
	if r.maxSize > 0 {
		if r.curSize+int64(len(msg)) > r.maxSize {
			return r.file, false, ErrFileSizeLimit
		}
	}
	return r.file, false, nil
}

func (r *TimeAndSizeRotator) RecordBytesWritten(n int) {
	if n <= 0 {
		return
	}
	r.curSize += int64(n)
}

func (r *TimeAndSizeRotator) DoRollover() (*os.File, error) {
	var err error

	if r.file != nil {
		if err = closeFile(&r.file); err != nil {
			return nil, err
		}
	}
	if !r.cfg.ConcurrentlyWrite {
		var parties []string
		if r.filename != "" {
			parties = append(parties, r.filename)
		}
		parties = append(parties, r.cfg.FileSuffix, r.rolloverTimeSuffix)
		dfn := filepath.Join(r.cfg.FileDir, strings.Join(parties, "."))

		if IsFileExist(r.filePath) {
			backupPath, err := availableBackupPath(dfn)
			if err != nil {
				return nil, err
			}
			err = os.Rename(r.filePath, backupPath)
			if err != nil {
				return nil, err
			}
		}
	}
	if r.backupCount > 0 {
		dir := r.cfg.FileDir
		filename := r.filename
		if !r.cfg.ConcurrentlyWrite {
			filename = filepath.Base(r.filePath)
		}
		var delFileList []string
		delFileList, err = collectRotatedBackupPaths(dir, filename, r.reCompile)
		if err != nil {
			return nil, err
		}
		if delFileLen := len(delFileList); delFileLen > r.backupCount {
			sort.Strings(delFileList)
			delFileList = delFileList[:delFileLen-r.backupCount]
			for _, filePath := range delFileList {
				err = os.Remove(filePath)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	// next rolloverAt
	r.rolloverAt = time.Now().Unix() + r.interval

	r.filePath = r.getNewFilepath()
	r.file, err = open(r.filePath)
	if err != nil {
		return nil, err
	}
	st, err := r.file.Stat()
	if err != nil {
		return nil, err
	}
	r.curSize = st.Size()
	return r.file, nil
}

func (r *TimeAndSizeRotator) getNewFilepath() string {
	var parties []string
	if r.filename != "" {
		parties = append(parties, r.filename)
	}
	r.rolloverTimeSuffix = time.Now().Format(r.suffixFmt)
	if r.cfg.ConcurrentlyWrite {
		parties = append(parties, r.rolloverTimeSuffix)
	}
	parties = append(parties, r.cfg.FileSuffix)
	return filepath.Join(r.cfg.FileDir, strings.Join(parties, "."))
}

func (r *TimeAndSizeRotator) Close() error {
	return closeFile(&r.file)
}

func IsFileExist(name string) bool {
	fileInfo, err := os.Stat(name)
	if fileInfo != nil && fileInfo.IsDir() {
		panic(fmt.Sprintf("This path '%s' is not a file path.", name))
	}
	return err == nil || os.IsExist(err)
}

func open(filepath string) (*os.File, error) {
	return os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
}

func mkdir(dir string) error {
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("make log directory %q: %w", dir, err)
	}
	return nil
}

func closeFile(file **os.File) error {
	if *file == nil {
		return nil
	}
	syncErr := (*file).Sync()
	closeErr := (*file).Close()
	*file = nil
	if syncErr != nil {
		return syncErr
	}
	return closeErr
}

func getFileModTime(file *os.File) (time.Time, error) {
	info, err := file.Stat()
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// collectRotatedBackupPaths lists files in dir whose names start with namePrefix+"." and suffix matches re.
func collectRotatedBackupPaths(dir, namePrefix string, re *regexp.Regexp) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fp := namePrefix + "."
	plen := len(fp)
	var out []string
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		fn := ent.Name()
		if len(fn) <= plen || fn[:plen] != fp {
			continue
		}
		suf := fn[plen:]
		matches := re.MatchString(suf)
		if !matches {
			if separator := strings.LastIndexByte(suf, '.'); separator >= 0 {
				if _, err := strconv.ParseUint(suf[separator+1:], 10, 64); err == nil {
					matches = re.MatchString(suf[:separator])
				}
			}
		}
		if matches {
			out = append(out, filepath.Join(dir, fn))
		}
	}
	return out, nil
}

func availableBackupPath(base string) (string, error) {
	entries, err := os.ReadDir(filepath.Dir(base))
	if err != nil {
		return "", err
	}
	filename := filepath.Base(base)
	var sequence uint64
	found := false
	for _, entry := range entries {
		if entry.Name() == filename {
			found = true
		}
		if strings.HasPrefix(entry.Name(), filename+".") {
			if previous, err := strconv.ParseUint(strings.TrimPrefix(entry.Name(), filename+"."), 10, 64); err == nil {
				found = true
				if previous > sequence {
					sequence = previous
				}
			}
		}
	}
	if !found {
		return base, nil
	}
	for sequence < ^uint64(0) {
		sequence++
		candidate := fmt.Sprintf("%s.%020d", base, sequence)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", errors.New("backup sequence exhausted")
}

func getRolloverSecond(interval int64) (rolloverAt int64) {
	if interval >= 60*60*24*7 || interval >= 60*60*24 {
		year, month, day := time.Now().Date()
		t := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		rolloverAt = t.Unix() + interval
	} else if interval >= 60*60 {
		rolloverAt = time.Now().Truncate(time.Hour).Unix() + interval
	} else if interval >= 60 {
		rolloverAt = time.Now().Truncate(time.Minute).Unix() + interval
	} else {
		rolloverAt = time.Now().Unix() + interval
	}
	return
}
