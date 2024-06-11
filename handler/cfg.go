package handler

/*
================== file ===================
*/

type FileHandlerConfig struct {
	FileDir       string
	FileName      string
	MaxFileSize   int64
	BackupCount   int
	BulkWriteSize int

	RotatorType       RotatorType
	Interval          int64  // unit: second. used in TimeRotator and TimeAndSizeRotator.
	TimeSuffixFmt     string // Time suffix format of file name:test.log.2024061104
	ReMatch           string
	FileSuffix        string
	ConcurrentlyWrite bool

	ErrCallback func(buf interface{}, err error)
}

func (c *FileHandlerConfig) WithFileDir(dir string) *FileHandlerConfig {
	c.FileDir = dir
	return c
}
func (c *FileHandlerConfig) WithFileName(name string) *FileHandlerConfig {
	c.FileName = name
	return c
}
func (c *FileHandlerConfig) WithFileSize(size int64) *FileHandlerConfig {
	c.MaxFileSize = size
	return c
}
func (c *FileHandlerConfig) WithBackupCount(n int) *FileHandlerConfig {
	c.BackupCount = n
	return c
}
func (c *FileHandlerConfig) WithBulkSize(size int) *FileHandlerConfig {
	c.BulkWriteSize = size
	return c
}
func (c *FileHandlerConfig) WithRotatorType(typ RotatorType) *FileHandlerConfig {
	c.RotatorType = typ
	return c
}
func (c *FileHandlerConfig) WithInterval(seconds int64) *FileHandlerConfig {
	c.Interval = seconds
	return c
}
func (c *FileHandlerConfig) WithTimeSuffixFmt(timeFormat string) *FileHandlerConfig {
	c.TimeSuffixFmt = timeFormat
	return c
}
func (c *FileHandlerConfig) WithReMatch(regexp string) *FileHandlerConfig {
	c.ReMatch = regexp
	return c
}
func (c *FileHandlerConfig) WithFileSuffix(suffix string) *FileHandlerConfig {
	c.FileSuffix = suffix
	return c
}
func (c *FileHandlerConfig) WithConcurrentlyWrite() *FileHandlerConfig {
	c.ConcurrentlyWrite = true
	return c
}
func (c *FileHandlerConfig) WithErrCallback(cb func(buf interface{}, err error)) *FileHandlerConfig {
	c.ErrCallback = cb
	return c
}

/*
================== Syslog ===================
*/

type SyslogHandlerConfig struct {
	Network  string
	Address  string
	Tag      string
	Priority int
}

func (c *SyslogHandlerConfig) WithNetwork(net string) *SyslogHandlerConfig {
	c.Network = net
	return c
}
func (c *SyslogHandlerConfig) WithAddress(addr string) *SyslogHandlerConfig {
	c.Address = addr
	return c
}
func (c *SyslogHandlerConfig) WithTag(tag string) *SyslogHandlerConfig {
	c.Tag = tag
	return c
}
func (c *SyslogHandlerConfig) WithPriority(priority int) *SyslogHandlerConfig {
	c.Priority = priority
	return c
}
