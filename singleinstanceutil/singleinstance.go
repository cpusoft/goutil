package singleinstanceutil

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"

	"github.com/cpusoft/goutil/belogs"
)

const (
	lockExclusive = syscall.LOCK_EX
	lockNonBlock  = syscall.LOCK_NB
)

// SingleInstanceLock 单实例锁，进程退出时自动释放
type SingleInstanceLock struct {
	file *os.File
	path string
}

// AcquireSingleInstance 获取单实例锁，失败返回 error（程序已运行）
func AcquireSingleInstance(lockPath string) (*SingleInstanceLock, error) {
	dir := filepath.Dir(lockPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		belogs.Error("AcquireSingleInstance(): MkdirAll fail, dir:", dir, err)
		return nil, err
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		belogs.Error("AcquireSingleInstance(): OpenFile fail, lockPath", lockPath, err)
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), lockExclusive|lockNonBlock); err != nil {
		_ = f.Close()
		belogs.Error("AcquireSingleInstance(): Flock fail, lockPath", lockPath, err)
		return nil, errors.New("application already running, forbid duplicate start")
	}

	return &SingleInstanceLock{file: f, path: lockPath}, nil
}

// Close 释放锁，通常放在 defer 中
func (l *SingleInstanceLock) Close() error {
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}
