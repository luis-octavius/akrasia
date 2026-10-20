//go:build unix

package web

import (
	"errors"
	"os"
	"syscall"
)

// Lock marks the database at dbPath as served, until release is called or
// the process ends (the kernel drops the lock, so a crash never leaves it
// behind). A second `serve` on the same database gets ErrRunning.
func Lock(dbPath string) (release func(), err error) {
	f, err := os.OpenFile(lockFile(dbPath), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrRunning
		}
		return nil, err
	}

	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}

// Running reports whether `serve` is using the database at dbPath.
func Running(dbPath string) bool {
	release, err := Lock(dbPath)
	if err != nil {
		return errors.Is(err, ErrRunning)
	}
	release()
	return false
}
