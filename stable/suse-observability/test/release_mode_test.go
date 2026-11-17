package test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var releaseModeReference = regexp.MustCompile(`\.Release\.Is(Upgrade|Install)\b`)

// Bitnami common only uses the release mode to reject an upgrade with an empty
// required password; the parent chart supplies those passwords.
const bitnamiUpgradePasswordCheck = "/charts/bitnami-common/templates/_errors.tpl"

// Tests render with `helm template` and do not repeat cases with --is-upgrade.
// That is only sound while no rendered template depends on the release mode.
func TestTemplatesDoNotDependOnReleaseMode(t *testing.T) {
	t.Parallel()

	require.NoError(t, filepath.WalkDir("../templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		requireNoReleaseModeReference(t, path, data)
		return nil
	}))

	archives, err := filepath.Glob("../charts/*.tgz")
	require.NoError(t, err)
	require.NotEmpty(t, archives, "build chart dependencies first")
	for _, archive := range archives {
		data, err := os.ReadFile(archive)
		require.NoError(t, err)
		scanChartArchive(t, archive, data)
	}
}

func scanChartArchive(t *testing.T, name string, archive []byte) {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	require.NoError(t, err)
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return
		}
		require.NoError(t, err)
		if header.Typeflag != tar.TypeReg {
			continue
		}
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		path := name + "!" + header.Name
		switch {
		case strings.HasSuffix(header.Name, ".tgz"):
			scanChartArchive(t, path, data)
		case strings.HasSuffix("/"+header.Name, bitnamiUpgradePasswordCheck):
		case strings.Contains(header.Name, "/templates/"):
			requireNoReleaseModeReference(t, path, data)
		}
	}
}

func requireNoReleaseModeReference(t *testing.T, path string, data []byte) {
	t.Helper()
	require.False(t, releaseModeReference.Match(data),
		"%s reads the release mode; restore --is-upgrade test cases for the affected resources", path)
}
