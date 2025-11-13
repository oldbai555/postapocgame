package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type CheckTimeToOpenNewFileFunc func(logName string, lastOpenFileTime *time.Time, isNeverOpenFile bool) (string, bool)

// OpenNewFileByByDateHour 按日期小时打开新文件的函数
func OpenNewFileByByDateHour(logName string, lastOpenFileTime *time.Time, isNeverOpenFile bool) (string, bool) {
	if isNeverOpenFile {
		return logName + time.Now().Format(".01-02.log"), true
	}

	lastOpenYear, lastOpenMonth, lastOpenDay := lastOpenFileTime.Date()

	now := time.Now()
	nowYear, nowMonth, nowDay := now.Date()

	if lastOpenDay != nowDay || lastOpenMonth != nowMonth || lastOpenYear != nowYear {
		return logName + time.Now().Format(".01-02.log"), true
	}

	return "", false
}

var (
	osStat      = os.Stat
	currentTime = time.Now
)

type FileLoggerWriter struct {
	fp                        *os.File
	baseDir                   string
	logName                   string // 日志文件名（不含扩展名）
	maxFileSize               int64
	lastCheckIsFullAt         int64
	isFileFull                bool
	checkFileFullIntervalSecs int64
	checkTimeToOpenNewFile    CheckTimeToOpenNewFileFunc
	openCurrentFileTime       *time.Time
	currentFileName           string
	bufCh                     chan []byte
	isFlushing                atomic.Bool
	flushSignCh               chan struct{}
	flushDoneSignCh           chan error
	mu                        sync.Mutex
	perm                      os.FileMode
	closed                    atomic.Bool
	stopCh                    chan struct{}
	wg                        sync.WaitGroup
	writeErrors               uint64 // 写入错误计数（用于监控）
}

func NewFileLoggerWriter(baseDir string, maxFileSize int64, checkFileFullIntervalSecs int64, checkTimeToOpenNewFile CheckTimeToOpenNewFileFunc, bufChanLen uint32, perm os.FileMode) *FileLoggerWriter {
	return &FileLoggerWriter{
		baseDir:                   strings.TrimRight(baseDir, "/"),
		maxFileSize:               maxFileSize,
		checkFileFullIntervalSecs: checkFileFullIntervalSecs,
		checkTimeToOpenNewFile:    checkTimeToOpenNewFile,
		bufCh:                     make(chan []byte, bufChanLen),
		flushSignCh:               make(chan struct{}, 1), // 缓冲1，避免阻塞
		flushDoneSignCh:           make(chan error, 1),
		stopCh:                    make(chan struct{}),
		perm:                      perm,
	}
}

// SetLogName 设置日志文件名
func (w *FileLoggerWriter) SetLogName(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.logName = name
}

func (w *FileLoggerWriter) checkFileIsFull() (bool, error) {
	if w.lastCheckIsFullAt != 0 && w.lastCheckIsFullAt+w.checkFileFullIntervalSecs > time.Now().Unix() {
		return w.isFileFull, nil
	}

	fileInfo, err := w.fp.Stat()
	if err != nil {
		return false, err
	}

	w.isFileFull = fileInfo.Size() >= w.maxFileSize
	w.lastCheckIsFullAt = time.Now().Unix()

	return w.isFileFull, nil
}

func (w *FileLoggerWriter) rotate() error {
	if err := w.close(); err != nil {
		return err
	}
	if err := w.openNew(); err != nil {
		return err
	}

	return nil
}

func (w *FileLoggerWriter) close() error {
	if w.fp == nil {
		return nil
	}
	err := w.fp.Close()
	w.fp = nil
	return err
}

func (w *FileLoggerWriter) openNew() error {
	err := os.MkdirAll(w.baseDir, w.perm)
	if err != nil {
		return err
	}

	name := w.baseDir + "/" + w.currentFileName
	mode := os.FileMode(0600)
	info, err := osStat(name)
	if err == nil {
		mode = info.Mode()
		newName := backUpName(name)
		if err := os.Rename(name, newName); err != nil {
			return err
		}
		if err := chown(name, info); err != nil {
			return err
		}
	}

	fp, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	openFileTime := time.Now()

	w.fp = fp
	w.openCurrentFileTime = &openFileTime
	w.isFileFull = false
	w.lastCheckIsFullAt = 0
	w.currentFileName = filepath.Base(name)
	return nil
}

func backUpName(name string) string {
	dir := filepath.Dir(name)
	fileName := filepath.Base(name)
	ext := filepath.Ext(fileName)
	prefix := fileName[:len(fileName)-len(ext)]
	timestamp := currentTime().Format(backupTimeFormat)
	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", prefix, timestamp, ext))
}

func (w *FileLoggerWriter) tryOpenNewFile() error {
	var err error
	w.mu.Lock()
	logName := w.logName
	w.mu.Unlock()

	fileName, ok := w.checkTimeToOpenNewFile(logName, w.openCurrentFileTime, w.openCurrentFileTime == nil)
	if !ok {
		if w.fp == nil {
			return errors.New("get first file name failed")
		}

		return nil
	}

	if w.fp == nil {
		if _, err = osStat(w.baseDir); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if err = os.MkdirAll(w.baseDir, w.perm); err != nil {
				return err
			}
		}
	}

	if w.fp, err = os.OpenFile(w.baseDir+"/"+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, w.perm); err != nil {
		return err
	}

	openFileTime := time.Now()
	w.openCurrentFileTime = &openFileTime
	w.isFileFull = false
	w.lastCheckIsFullAt = 0
	w.currentFileName = fileName

	return nil
}

func (w *FileLoggerWriter) Flush() error {
	if w.closed.Load() {
		return ErrWriterClosed
	}

	w.isFlushing.Store(true)
	select {
	case w.flushSignCh <- struct{}{}:
		return <-w.flushDoneSignCh
	case <-w.stopCh:
		return ErrWriterClosed
	}
}

func (w *FileLoggerWriter) finishFlush(err error) {
	w.isFlushing.Store(false)
	w.flushDoneSignCh <- err
}

func (w *FileLoggerWriter) isFlushingNow() bool {
	return w.isFlushing.Load()
}

func (w *FileLoggerWriter) Write(logContent []byte) (n int, err error) {
	if w.closed.Load() {
		return 0, ErrWriterClosed
	}

	select {
	case w.bufCh <- logContent:
		return len(logContent), nil
	case <-w.stopCh:
		return 0, ErrWriterClosed
	default:
		// 缓冲满时丢弃，避免阻塞主线程（记录错误计数）
		atomic.AddUint64(&w.writeErrors, 1)
		return 0, ErrBufferFull
	}
}

// Loop 主循环（带错误恢复和优雅关闭）
func (w *FileLoggerWriter) Loop() error {
	defer func() {
		// 确保关闭文件
		w.mu.Lock()
		if w.fp != nil {
			w.fp.Sync()
			w.fp.Close()
			w.fp = nil
		}
		w.mu.Unlock()
	}()

	doWriteMoreAsPossible := func(buf []byte) error {
		// 批量读取尽可能多的日志
		for {
			var moreBuf []byte
			select {
			case moreBuf = <-w.bufCh:
				buf = append(buf, moreBuf...)
			default:
			}

			if moreBuf == nil {
				break
			}
		}

		if len(buf) == 0 {
			return nil
		}

		// 尝试打开新文件（如果需要）
		if err := w.tryOpenNewFile(); err != nil {
			// 文件打开失败时，尝试降级处理
			return w.handleWriteError(err)
		}

		// 检查文件是否已满
		if isFull, err := w.checkFileIsFull(); err != nil {
			return w.handleWriteError(err)
		} else if isFull {
			if err := w.rotate(); err != nil {
				return w.handleWriteError(err)
			}
		}

		// 写入文件
		return w.writeToFile(buf)
	}

	for {
		select {
		case buf := <-w.bufCh:
			if err := doWriteMoreAsPossible(buf); err != nil {
				// 错误已处理，继续运行
				continue
			}
		case <-w.flushSignCh:
			if err := doWriteMoreAsPossible([]byte{}); err != nil {
				w.finishFlush(err)
				continue
			}
			w.mu.Lock()
			if w.fp != nil {
				if err := w.fp.Sync(); err != nil {
					w.mu.Unlock()
					w.finishFlush(err)
					continue
				}
			}
			w.mu.Unlock()
			w.finishFlush(nil)
		case <-w.stopCh:
			// 优雅关闭：处理剩余日志
			for {
				select {
				case buf := <-w.bufCh:
					doWriteMoreAsPossible(buf)
				default:
					// 处理完所有日志后退出
					return nil
				}
			}
		}
	}
}

// writeToFile 写入文件（带重试）
func (w *FileLoggerWriter) writeToFile(buf []byte) error {
	w.mu.Lock()
	fp := w.fp
	w.mu.Unlock()

	if fp == nil {
		return errors.New("file not opened")
	}

	bufLen := len(buf)
	var totalWrittenBytes int
	maxRetries := 3

	for retries := 0; retries < maxRetries; retries++ {
		n, err := fp.Write(buf[totalWrittenBytes:])
		if err != nil {
			if retries < maxRetries-1 {
				// 重试前等待一小段时间
				time.Sleep(time.Millisecond * 10)
				continue
			}
			atomic.AddUint64(&w.writeErrors, 1)
			return err
		}
		totalWrittenBytes += n
		if totalWrittenBytes >= bufLen {
			return nil
		}
	}

	return nil
}

// handleWriteError 处理写入错误（降级策略）
func (w *FileLoggerWriter) handleWriteError(err error) error {
	atomic.AddUint64(&w.writeErrors, 1)
	// 可以在这里添加降级逻辑，比如输出到 stderr
	// 当前实现：记录错误但继续运行
	return nil
}

// Close 优雅关闭写入器
func (w *FileLoggerWriter) Close() error {
	if !w.closed.CompareAndSwap(false, true) {
		return nil // 已经关闭
	}

	close(w.stopCh)
	w.wg.Wait()

	w.mu.Lock()
	defer w.mu.Unlock()
	if w.fp != nil {
		w.fp.Sync()
		err := w.fp.Close()
		w.fp = nil
		return err
	}
	return nil
}

// GetWriteErrors 获取写入错误计数
func (w *FileLoggerWriter) GetWriteErrors() uint64 {
	return atomic.LoadUint64(&w.writeErrors)
}
