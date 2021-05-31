package datastore

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestSegment(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-segment")
	if err != nil {
		t.Fatal("Creating db store error")
	}

	defer os.RemoveAll(dir)

	name := strconv.FormatInt(time.Now().UnixNano(), 10)

	file, err := openFile(dir, name, os.O_APPEND|os.O_RDWR|os.O_CREATE)
	if err != nil {
		t.Error(err)
	}

	seg := NewSegment(file)

	pairs := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	t.Run("put/get", func(t *testing.T) {
		for key, val := range pairs {
			if err = seg.Write(key, val); err != nil {
				t.Errorf("Cannot put %s: %s", key, err)
			}

			var value []byte

			if value, err = seg.Get(key); err != nil {
				t.Errorf("Cannot get %s: %s", key, err)
			}

			if !bytes.Equal(value, val) {
				t.Errorf("Bad value returned expected %s, got %s", val, value)
			}
		}
	})

	t.Run("segment recovery", func(t *testing.T) {
		if err = seg.Close(); err != nil {
			t.Fatal(err)
		}

		file, err = openFile(dir, name, os.O_APPEND|os.O_RDWR)
		if err != nil {
			t.Error(err)
		}

		seg = NewSegment(file)
		if err = seg.Recover(); err != nil && !errors.Is(err, io.EOF) {
			t.Log(err)
		}

		for key, val := range pairs {
			var value []byte

			value, err = seg.Get(key)
			if err != nil {
				t.Errorf("Cannot get %s: %s", key, err)
			}

			if !bytes.Equal(value, val) {
				t.Errorf("Bad value returned expected %s, got %s", val, value)
			}
		}
	})
}
