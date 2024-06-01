//go:build !windows
// +build !windows

package handler

// import "C"
import (
	"os"
	"sync"
)

var stderrFile *os.File
var once sync.Once

func rewriteStderr(fileDir, filePrefix string) {
	// once.Do(func() {
	//	stderrFilepath := filepath.Join(fileDir, fmt.Sprintf("%s_stderr.log", filePrefix))
	//	file, err := os.OpenFile(stderrFilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//	if err != nil {
	//		println(err)
	//		return
	//	}
	//	// Save the file handle to a global variable to avoid being reclaimed by GC
	//	stderrFile = file
	//
	//	if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
	//		println(err)
	//		return
	//	}
	//	// Close file descriptors before memory reclamation
	//	runtime.SetFinalizer(stderrFile, func(fd *os.File) {
	//		_ = fd.Close()
	//	})
	// })
}
