// BEGIN: 3d2f5a8c5b2c
package device

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetDirectories(t *testing.T) {
	// create temporary directory for testing
	tempDir := t.TempDir()

	prefix := "fdpm_"

	// create subdirectories with filter prefix
	subDir1 := filepath.Join(tempDir, fmt.Sprintf("%sdir1", prefix))
	subDir2 := filepath.Join(tempDir, fmt.Sprintf("%sdir2", prefix))
	subDir3 := filepath.Join(tempDir, fmt.Sprintf("%sdir3", prefix))
	os.Mkdir(subDir1, 0755)
	os.Mkdir(subDir2, 0755)
	os.Mkdir(subDir3, 0755)

	// test with filter prefix
	subDirs, err := GetDirectories(tempDir, prefix)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check if subdirectories were found
	if len(subDirs) != 3 {
		t.Errorf("expected 3 subdirectories, got %d", len(subDirs))
	}
}

// END: 3d2f5a8c5b2c
