package mtimehash

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcess(t *testing.T) {
	fileToContent := map[string]string{
		"a.txt": "aaa",
		"b.txt": "bbb",
		"c.txt": "aaa",
	}
	files := setupFiles(t, fileToContent)

	t.Run("happy path", func(t *testing.T) {
		require.NoError(t, Process(slices.Values(files), 1000000000))

		mtimes := getMtimes(t, files)
		assert.Equal(t, map[string]int64{
			"a.txt": 259627185,
			"b.txt": 613142970,
			"c.txt": 259627185,
		}, mtimes)
	})

	t.Run("low maxUnixTime", func(t *testing.T) {
		require.NoError(t, Process(slices.Values(files), 2))

		mtimes := getMtimes(t, files)
		assert.Equal(t, map[string]int64{
			"a.txt": 1,
			"b.txt": 0,
			"c.txt": 1,
		}, mtimes)
	})

	t.Run("errors", func(t *testing.T) {
		var badFiles []string

		badFiles = append(badFiles, "nonexistent.txt")

		dirPath := filepath.Join(t.TempDir(), "dir")
		require.NoError(t, os.Mkdir(dirPath, 0o777))
		badFiles = append(badFiles, dirPath)

		nonReadableFilePath := filepath.Join(t.TempDir(), "non-readable.txt")
		require.NoError(t, os.WriteFile(nonReadableFilePath, []byte("non-readable"), 0o000))
		badFiles = append(badFiles, nonReadableFilePath)

		err := Process(slices.Values(slices.Concat(badFiles, files)), 1000000000)
		assert.Error(t, err)

		mtimes := getMtimes(t, files)
		assert.Equal(t, map[string]int64{
			"a.txt": 259627185,
			"b.txt": 613142970,
			"c.txt": 259627185,
		}, mtimes)
	})
}

func setupFiles(t *testing.T, files map[string]string) []string {
	t.Helper()
	tempDir := t.TempDir()
	var filePaths []string
	for name, content := range files {
		filePath := filepath.Join(tempDir, name)
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o666))
		filePaths = append(filePaths, filePath)
	}
	return filePaths
}

func getMtimes(t *testing.T, files []string) map[string]int64 {
	t.Helper()
	mtimes := make(map[string]int64)
	for _, file := range files {
		s, err := os.Stat(file)
		require.NoError(t, err)
		mtimes[path.Base(file)] = s.ModTime().Unix()
	}
	return mtimes
}
