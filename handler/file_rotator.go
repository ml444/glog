package handler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ml444/glog/util"
)

type IRotator interface {
	NeedRollover(msg []byte) (*os.File, bool, error)
	DoRollover() (*os.File, error)
	Close() error
}

func NewRotator(cfg *FileHandlerConfig) (IRotator, error) {
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
}

func NewSizeRotator(cfg *FileHandlerConfig) (*SizeRotator, error) {
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
	var err error
	file := r.file
	if file == nil {
		file, err = open(r.filePath)
		if err != nil {
			return nil, false, err
		}
		r.file = file
	}
	if r.maxSize > 0 {
		var size int64
		size, err = file.Seek(0, io.SeekEnd)
		if err == nil && size+int64(len(msg)) >= r.maxSize {
			return file, true, nil
		}
	}
	return file, false, err
}

func (r *SizeRotator) DoRollover() (*os.File, error) {
	if r.file != nil {
		_ = r.file.Sync()
		_ = r.file.Close()
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
	return f, nil
}

func (r *SizeRotator) Close() error {
	if r.file == nil {
		return errors.New("file not open")
	}
	return r.file.Sync()
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
	if r.interval <= 0 {
		r.interval = 60 * 60 // default 1 hour
	}
	r.rolloverAt = getRolloverSecond(r.interval)
	r.reCompile = regexp.MustCompile(r.cfg.ReMatch)

	err := mkdir(r.cfg.FileDir)
	if err != nil {
		panic(err)
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
	t := time.Now().Unix()
	if t >= r.rolloverAt {
		return r.file, true, nil
	}
	return r.file, false, nil
}

func (r *TimeRotator) DoRollover() (*os.File, error) {
	if r.file != nil {
		_ = r.file.Sync()
		_ = r.file.Close()
	}
	curTime := time.Now()
	suffixTime := curTime.Format(r.cfg.TimeSuffixFmt)
	dfn := fmt.Sprintf("%s.%s", r.filePath, suffixTime)
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
	if backupCount := r.cfg.BackupCount; backupCount > 0 {
		dir, filename := filepath.Split(r.filePath)
		var delFileList []string
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			filePreFix := filename + "."
			pLen := len(filePreFix)
			if fn := info.Name(); fn[:pLen] == filePreFix {
				fileSuffix := fn[pLen:]
				if r.reCompile.MatchString(fileSuffix) {
					delFileList = append(delFileList, filepath.Join(path, fn))
				}
			}
			return nil
		})
		if delLen := len(delFileList); delLen > backupCount {
			sort.Strings(delFileList)
			delFileList = delFileList[:delLen-backupCount]
		}
		for _, filePath := range delFileList {
			err := os.Remove(filePath)
			if err != nil {
				return nil, err
			}
		}
	}

	// next rolloverAt
	r.rolloverAt = curTime.Unix() + r.interval

	f, err := open(r.filePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *TimeRotator) Close() error {
	if r.file == nil {
		return errors.New("file not open")
	}
	return r.file.Sync()
}

type TimeAndSizeRotator struct {
	cfg                *FileHandlerConfig
	file               *os.File
	filename           string
	filePath           string
	maxSize            int64
	backupCount        int
	interval           int64
	suffixFmt          string
	rolloverAt         int64
	rolloverTimeSuffix string
	reCompile          *regexp.Regexp
}

func NewTimeAndSizeRotator(cfg *FileHandlerConfig) (*TimeAndSizeRotator, error) {
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
	r.reCompile = regexp.MustCompile(r.cfg.ReMatch)

	err := mkdir(r.cfg.FileDir)
	if err != nil {
		panic(err)
	}

	r.filePath = r.getNewFilepath()
	return nil
}

func (r *TimeAndSizeRotator) NeedRollover(msg []byte) (*os.File, bool, error) {
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
	}
	t := time.Now().Unix()
	if t >= r.rolloverAt {
		return r.file, true, nil
	}
	if r.maxSize > 0 {
		size, err := r.file.Seek(0, io.SeekEnd)
		if err != nil {
			return r.file, false, err
		} else {
			if size+int64(len(msg)) >= r.maxSize {
				return r.file, false, errors.New("maximum file size limit")
			} else {
				return r.file, false, nil
			}
		}
	}
	return r.file, false, nil
}

func (r *TimeAndSizeRotator) DoRollover() (*os.File, error) {
	var err error

	if r.file != nil {
		err = r.file.Sync()
		if err != nil {
			return nil, err
		}
		err = r.file.Close()
		if err != nil {
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

		if IsFileExist(dfn) {
			err = os.Remove(dfn)
			if err != nil {
				return nil, err
			}
		}
		if IsFileExist(r.filePath) {
			err = os.Rename(r.filePath, dfn)
			if err != nil {
				return nil, err
			}
		}
	}
	if r.backupCount > 0 {
		dir := r.cfg.FileDir
		filename := r.filename
		filePreFix := filename + "."
		pLen := len(filePreFix)
		var delFileList []string
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info == nil || info.IsDir() {
				return nil
			}
			if fn := info.Name(); strings.HasPrefix(fn, filePreFix) {
				fileSuffix := fn[pLen:]
				if r.reCompile.MatchString(fileSuffix) {
					delFileList = append(delFileList, path)
				}
			}
			return nil
		})
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
	if r.file == nil {
		return errors.New("file not open")
	}
	return r.file.Sync()
}

func IsFileExist(name string) bool {
	fileInfo, err := os.Stat(name)
	if fileInfo != nil && fileInfo.IsDir() {
		panic(fmt.Sprintf("This path '%s' is not a file path.", name))
	}
	return err == nil || os.IsExist(err)
}

func open(filepath string) (*os.File, error) {
	old := util.UMask(0)
	defer util.UMask(old)
	return os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
}

func mkdir(dir string) error {
	if dir == "" {
		dir = "."
	}
	if !strings.HasPrefix(dir, ".") {
		old := util.UMask(0)
		defer util.UMask(old)
		err := os.MkdirAll(dir, 0o777)
		if err != nil {
			println(fmt.Sprintf("make dir fail, dir %s, err %s\n", dir, err))
			return err
		}
	}
	return nil
}

func getFileModTime(file *os.File) (time.Time, error) {
	info, err := file.Stat()
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
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
