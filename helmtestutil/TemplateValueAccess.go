package helmtestutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// RequireTemplateValueAccessViaHelper rejects ordinary dotted Values access
// outside the default helper file. This guards accidental bypasses of parent
// configuration; it is not a general template parser or data-flow analysis.
func RequireTemplateValueAccessViaHelper(t *testing.T, templatesDir, helperFile, pattern, helperName string) {
	t.Helper()
	directAccess := regexp.MustCompile(pattern)
	comments := regexp.MustCompile(`(?s)\{\{-?\s*/\*.*?\*/\s*-?\}\}`)
	err := filepath.WalkDir(templatesDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path == filepath.Join(templatesDir, helperFile) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, access := range directAccess.FindAll(comments.ReplaceAll(content, nil), -1) {
			t.Errorf("%s: direct %s access bypasses parent configuration; use %s instead", path, access, helperName)
		}
		return nil
	})
	require.NoError(t, err)
}
