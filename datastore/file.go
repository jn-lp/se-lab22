package datastore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func openFile(dir, fileName string) (file *os.File, size int64, err error) {
	var fi os.FileInfo

	if fi, err = os.Stat(dir); err == nil {
		if !fi.IsDir() {
			return nil, 0, fmt.Errorf("%s is not a directory", dir)
		}
	} else if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, 0, err
		}
	} else {
		return nil, 0, err
	}

	outPath := filepath.Join(dir, fileName)

	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, 0, err
	}

	fSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		_ = f.Close()

		return nil, 0, err
	}

	return f, fSize, nil
}
