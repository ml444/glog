package handlers

import (
	"errors"
	"fmt"
	"github.com/ml444/glog/config"
	"github.com/ml444/glog/util"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type IRotator interface {
	//Init()
	NeedRollover(msg []byte) (*os.File, bool, error)
	DoRollover() (*os.File, error)
	Close() error
}

func GetRotator4Config(rotatorCfg *config.FileRotatorConfig) (IRotator, error) {
	switch rotatorCfg.Type {
	case config.FileRotatorTypeSize:
		return NewSizeRotator(rotatorCfg)
	case config.FileRotatorTypeTime:
		return NewTimeRotator(rotatorCfg)
	case config.FileRotatorTypeTimeAndSize:
		return NewTimeAndSizeRotator(rotatorCfg)
	default:
		return NewTimeAndSizeRotator(rotatorCfg)
	}
}

type SizeRotator struct {
	file        *os.File
	filePath    string
	maxSize     int64
	backupCount int
}

func NewSizeRotator(rotatorCfg *config.FileRotatorConfig) (*SizeRotator, error) {
	return &SizeRotator{}, nil
}

func (r *SizeRotator) NeedRollover(msg []byte) (*os.File, bool, error) {
	var err error
	if r.file == nil {
		r.file, err = open(r.filePath)
		if err != nil {
			fmt.Println(err)
			return nil, false, err
		}
	}
	if r.maxSize > 0 {
		var size int64
		size, err = r.file.Seek(0, io.SeekEnd)
		if err == nil && size+int64(len(msg)) >= r.maxSize {
			return r.file, true, nil
		}
	}
	return r.file, false, err
}

func (r *SizeRotator) DoRollover() (*os.File, error) {
	if r.file != nil {
		r.file.Close()
	}
	if r.backupCount > 0 {
		for i := r.backupCount - 1; i <= 0; i-- {
			sfn := fmt.Sprintf("%s.%d", r.filePath, i)
			dfn := fmt.Sprintf("%s.%d", r.filePath, i+1)
			if IsFileExist(sfn) {
				if IsFileExist(dfn) {
					os.Remove(dfn)
				}
				err := os.Rename(sfn, dfn)
				if err != nil {
					panic(err)
				}
			}
		}
		dfn := fmt.Sprintf("%s.1", r.filePath)
		if IsFileExist(dfn) {
			os.Remove(dfn)
		}
		if IsFileExist(r.filePath) {
			err := os.Rename(r.filePath, dfn)
			if err != nil {
				panic(err)
			}
		}
	}

	f, err := open(r.filePath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *SizeRotator) Close() error {
	if r.file == nil {
		return errors.New("file not open")
	}
	return r.file.Sync()
}


type TimeRotator struct {
	cfg      *config.FileRotatorConfig
	file     *os.File
	filePath string
	//when         uint8
	intervalStep int64
	interval     int64
	suffixFmt    string
	reMatch      string
	backupCount  int
	rolloverAt   int64
	reCompile    *regexp.Regexp
}

func NewTimeRotator(rotatorCfg *config.FileRotatorConfig) (*TimeRotator, error) {
	r := &TimeRotator{
		cfg:      rotatorCfg,
		filePath: "",
		//when:         0,
		intervalStep: 0,
		interval:     0,
		suffixFmt:    "",
		reMatch:      "",
		backupCount:  0,
		rolloverAt:   0,
		reCompile:    nil,
	}
	err := r.Init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *TimeRotator) Init() error {
	if r.cfg == nil {
		return errors.New("rotator config is nil")
	}
	switch r.cfg.When {
	case config.FileRotatorWhenSecond:
		r.interval = 1 // one second
		r.suffixFmt = "2006-01-02_15-04-05"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}-\\d{2}(\\.\\w+)?$"
	case config.FileRotatorWhenMinute:
		r.interval = 60 // one minute
		r.suffixFmt = "2006-01-02_15-04"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}(\\.\\w+)?$"
	case config.FileRotatorWhenHour:
		r.interval = 60 * 60
		r.suffixFmt = "2006-01-02_15"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}(\\.\\w+)?$"
	case config.FileRotatorWhenDay:
		r.interval = 60 * 60 * 24
		r.suffixFmt = "2006-01-02"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}(\\.\\w+)?$"
	//case "W":
	//	r.interval = 60 * 60 * 24 * 7 // one minute
	//	r.suffixFmt = "2006-01-02"
	//	r.reMatch = "^\\d{4}-\\d{2}-\\d{2}(\\.\\w+)?$"
	default:
		panic(fmt.Sprintf("Invalid rollover interval specified: %d", r.cfg.When))
	}
	reCompile, err := regexp.Compile(r.reMatch)
	if err != nil {
		panic(err)
	}
	r.reCompile = reCompile
	if r.intervalStep != 0 {
		r.interval = r.interval * r.intervalStep
	}

	err = mkdir(r.cfg.FileDir)
	if err != nil {
		panic(err)
	}

	r.cfg.FileName = removeSuffix(r.cfg.FileName, ".log")
	r.filePath = filepath.Join(r.cfg.FileDir, r.cfg.FileName)
	return nil
}

func (r *TimeRotator) NeedRollover(msg []byte) (*os.File, bool, error) {
	t := time.Now().Unix()
	if t >= r.rolloverAt {
		return r.file, true, nil
	}
	return r.file, false, nil
}

// computeRollover: Work out the rollover time based on the specified time.
func (r *TimeRotator) computeRollover(currentTime int64) int64 {
	result := currentTime + r.interval
	return result
}

func (r *TimeRotator) DoRollover() (*os.File, error) {
	if r.file != nil {
		// TODO Flush()
		r.file.Close()
	}
	curTime := time.Now()
	suffixTime := curTime.Format(r.suffixFmt)
	dfn := fmt.Sprintf("%s.%s", r.filePath, suffixTime)
	if IsFileExist(dfn) {
		os.Remove(dfn)
	}
	if IsFileExist(r.filePath) {
		err := os.Rename(r.filePath, dfn)
		if err != nil {
			panic(err)
		}
	}
	if r.backupCount > 0 {
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
		if len(delFileList) > r.backupCount {
			sort.Strings(delFileList)
			delFileList = delFileList[:len(delFileList)-r.backupCount]
		}
		for _, filePath := range delFileList {
			os.Remove(filePath)
		}
	}
	newRolloverAt := r.computeRollover(curTime.Unix())
	r.rolloverAt = newRolloverAt

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
	cfg         *config.FileRotatorConfig
	file        *os.File
	filePath    string
	maxSize     int64
	backupCount int

	intervalStep int64
	interval     int64
	suffixFmt    string
	reMatch      string
	rolloverAt   int64
	reCompile    *regexp.Regexp
}

func NewTimeAndSizeRotator(rotatorCfg *config.FileRotatorConfig) (*TimeAndSizeRotator, error) {
	r := &TimeAndSizeRotator{
		cfg:          rotatorCfg,
		filePath:     "",
		maxSize:      rotatorCfg.MaxFileSize,
		backupCount:  rotatorCfg.BackupCount,
		intervalStep: rotatorCfg.IntervalStep,
		interval:     0,
		suffixFmt:    "",
		reMatch:      "",
		rolloverAt:   0,
		reCompile:    nil,
	}
	err := r.Init()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *TimeAndSizeRotator) Init() error {
	if r.cfg == nil {
		return errors.New("rotator config is nil")
	}
	switch r.cfg.When {
	case config.FileRotatorWhenSecond:
		r.interval = 1 // one second
		r.suffixFmt = "2006-01-02_15-04-05"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}-\\d{2}(\\.\\w+)?$"
	case config.FileRotatorWhenMinute:
		r.interval = 60 // one minute
		r.suffixFmt = "2006-01-02_15-04"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}-\\d{2}(\\.\\w+)?$"
	case config.FileRotatorWhenHour:
		r.interval = 60 * 60
		r.suffixFmt = "2006-01-02_15"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}_\\d{2}(\\.\\w+)?$"
	case config.FileRotatorWhenDay:
		r.interval = 60 * 60 * 24
		r.suffixFmt = "2006-01-02"
		r.reMatch = "^\\d{4}-\\d{2}-\\d{2}(\\.\\w+)?$"
	//case "W":
	//	r.interval = 60 * 60 * 24 * 7 // one minute
	//	r.suffixFmt = "2006-01-02"
	//	r.reMatch = "^\\d{4}-\\d{2}-\\d{2}(\\.\\w+)?$"
	default:
		panic(fmt.Sprintf("Invalid rollover interval specified: %d", r.cfg.When))
	}
	reCompile := regexp.MustCompile(r.reMatch)
	r.reCompile = reCompile
	if r.intervalStep != 0 {
		r.interval = r.interval * r.intervalStep
	}

	err := mkdir(r.cfg.FileDir)
	if err != nil {
		panic(err)
	}

	r.cfg.FileName = removeSuffix(r.cfg.FileName, ".log")
	r.filePath = filepath.Join(r.cfg.FileDir, r.cfg.FileName)
	r.file, err = open(r.filePath)
	if err != nil {
		return err
	}
	return nil
}
func (r *TimeAndSizeRotator) NeedRollover(msg []byte) (*os.File, bool, error) {
	if r.file == nil {
		var err error
		r.file, err = open(r.filePath)
		if err != nil {
			println(err)
			return r.file, false, err
		}
	}
	t := time.Now().Unix()
	if t >= r.rolloverAt {
		return r.file, true, nil
	}
	if r.maxSize > 0 {
		size, err := r.file.Seek(0, io.SeekEnd)
		if err == nil {
			if size+int64(len(msg)) >= r.maxSize {
				return r.file, false, errors.New("file has full")
			} else {
				return r.file, false, nil
			}
		} else {
			return r.file, false, err
		}
	}
	return r.file, false, nil
}

// computeRollover: Work out the rollover time based on the specified time.
func (r *TimeAndSizeRotator) computeRollover(currentTime int64) int64 {
	result := currentTime + r.interval
	return result
}

func (r *TimeAndSizeRotator) DoRollover() (*os.File, error) {
	var err error

	if r.file != nil {
		_ = r.file.Sync()
		_ = r.file.Close()
	}
	curTime := time.Now()
	suffixTime := curTime.Format(r.suffixFmt)
	dfn := fmt.Sprintf("%s.%s", r.filePath, suffixTime)
	if IsFileExist(dfn) {
		os.Remove(dfn)
	}
	if IsFileExist(r.filePath) {
		err := os.Rename(r.filePath, dfn)
		if err != nil {
			panic(err)
		}
	}
	if r.backupCount > 0 {
		dir, filename := filepath.Split(r.filePath)
		var delFileList []string
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if info == nil || info.IsDir() {
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
		if len(delFileList) > r.backupCount {
			sort.Strings(delFileList)
			delFileList = delFileList[:len(delFileList)-r.backupCount]
		}
		for _, filePath := range delFileList {
			os.Remove(filePath)
		}
	}
	newRolloverAt := r.computeRollover(curTime.Unix())
	r.rolloverAt = newRolloverAt

	r.file, err = open(r.filePath)
	if err != nil {
		return nil, err
	}
	return r.file, nil
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
	//old := util.UMask(0)
	//defer util.UMask(old)
	return os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
}

func mkdir(dir string) error {
	if dir == "" {
		dir = "."
	}
	if !strings.HasPrefix(dir, ".") {
		old := util.UMask(0)
		defer util.UMask(old)
		err := os.MkdirAll(dir, 0777)
		if err != nil {
			println(fmt.Sprintf("make dir fail, dir %s, err %s\n", dir, err))
			return err
		}
	}
	return nil
}
func removeSuffix(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[0 : len(s)-len(suffix)]
	}
	return s
}
