package datastore

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestDb_Put(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-DB")
	if err != nil {
		t.Fatal(err)
	}

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(dir)

	testDB, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	pairs := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	t.Run("put/get", func(t *testing.T) {
		for key, val := range pairs {
			if err = testDB.Put(key, val); err != nil {
				t.Errorf("Cannot put %s: %s", key, err)
			}

			var value []byte

			if value, err = testDB.Get(key); err != nil {
				t.Errorf("Cannot get %s: %s", key, err)
			}

			if !bytes.Equal(value, val) {
				t.Errorf("Bad value returned expected %s, got %s", val, value)
			}
		}
	})

	t.Run("new DB process", func(t *testing.T) {
		if err = testDB.Close(); err != nil {
			t.Fatal(err)
		}

		testDB, err = New(dir)
		if err != nil {
			t.Fatal(err)
		}

		for key, val := range pairs {
			var value []byte

			value, err = testDB.Get(key)
			if err != nil {
				t.Errorf("Cannot put %s: %s", key, err)
			}

			if !bytes.Equal(value, val) {
				t.Errorf("Bad value returned expected %s, got %s", val, value)
			}
		}
	})
}
