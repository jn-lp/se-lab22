package datastore

import (
	"fmt"
	"os"
	"path/filepath"
)

func openFile(dir, fileName string, flag int) (file *os.File, err error) {
	var fi os.FileInfo

	if fi, err = os.Stat(dir); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", dir)
		}
	} else if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	outPath := filepath.Join(dir, fileName)

	f, err := os.OpenFile(outPath, flag, 0o644)
	if err != nil {
		return nil, err
	}

	return f, nil
}
