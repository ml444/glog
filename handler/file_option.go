package handler

type FileOpt func(cfg *FileConfig)

func WithFileDir(dir string) FileOpt {
	return func(cfg *FileConfig) {
		cfg.FileDir = dir
	}
}

func WithFileName(fileName string) FileOpt {
	return func(cfg *FileConfig) {
		cfg.FileName = fileName
	}
}

func WithFileMaxFileSize(maxFileSize int64) FileOpt {
	return func(cfg *FileConfig) {
		cfg.MaxFileSize = maxFileSize
	}
}

func WithFileBackupCount(backupCount int) FileOpt {
	return func(cfg *FileConfig) {
		cfg.BackupCount = backupCount
	}
}

func WithFileBulkWriteSize(bulkWriteSize int) FileOpt {
	return func(cfg *FileConfig) {
		cfg.BulkWriteSize = bulkWriteSize
	}
}

func WithFileRotatorType(rotatorType RotatorType) FileOpt {
	return func(cfg *FileConfig) {
		cfg.RotatorType = rotatorType
	}
}

func WithFileInterval(interval int64) FileOpt {
	return func(cfg *FileConfig) {
		cfg.Interval = interval
	}
}

func WithFileTimeSuffixFmt(timeSuffixFmt string) FileOpt {
	return func(cfg *FileConfig) {
		cfg.TimeSuffixFmt = timeSuffixFmt
	}
}

func WithFileReMatch(reMatch string) FileOpt {
	return func(cfg *FileConfig) {
		cfg.ReMatch = reMatch
	}
}

func WithFileFileSuffix(fileSuffix string) FileOpt {
	return func(cfg *FileConfig) {
		cfg.FileSuffix = fileSuffix
	}
}

func WithFileMultiProcessWrite(multiProcessWrite bool) FileOpt {
	return func(cfg *FileConfig) {
		cfg.MultiProcessWrite = multiProcessWrite
	}
}

func WithFileErrCallback(errCallback func(buf []byte, err error)) FileOpt {
	return func(cfg *FileConfig) {
		cfg.ErrCallback = errCallback
	}
}
