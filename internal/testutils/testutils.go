package testutils

import "path/filepath"

// GetTestDataFilePath returns the full path of a file in the testdata/ dir
func GetTestDataFilePath(testFile string) (string, error) {
	testdataPath, err := filepath.Abs(filepath.FromSlash("../testdata"))
	if err != nil {
		return "", err
	}

	return filepath.Join(testdataPath, testFile), nil
}
