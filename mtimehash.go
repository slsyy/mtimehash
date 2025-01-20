package mtimehash

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/sourcegraph/conc/pool"
)

// Process input files and update their modification time based on the hash of their content.
// The modification time is set to the hash modulo maxUnixTime.
func Process(input iter.Seq[string], maxUnixTime int64) error {
	p := pool.New().WithErrors().WithMaxGoroutines(runtime.GOMAXPROCS(0))
	for filePath := range input {
		p.Go(func() error {
			err := updateMtime(filePath, maxUnixTime)
			if err != nil {
				slog.Default().Error("failed to process file", "file", filePath, "err", err)
			}
			return err
		})
	}
	return p.Wait()
}

// hashToTime converts a hash to a timestamp
func hashToTime(h64 uint64, maxUnixTime int64) time.Time {
	sec := h64 % uint64(maxUnixTime)
	return time.Unix(int64(sec), 0)
}

// updateMtime updates the file's modification time
func updateMtime(filePath string, maxUnixTime int64) error {
	logger := slog.Default()

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}
	if !s.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file, got %s", filePath, s.Mode())
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hash file: %w", err)
	}
	h64 := binary.BigEndian.Uint64(h.Sum(nil)[:8]) // take first 8 bytes of the hash
	mtime := hashToTime(h64, maxUnixTime)

	if err := os.Chtimes(filePath, time.Time{}, mtime); err != nil {
		return fmt.Errorf("set mtime: %w", err)
	}

	logger.Debug("updated file modification time", "file", filePath, "mtime", mtime)
	return nil
}
