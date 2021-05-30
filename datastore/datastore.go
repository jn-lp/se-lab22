package datastore

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const bufSize = 8192

type Datastore interface {
	WithOptions(...DSOption) Datastore
	Options() DSOptions

	Recover() error

	Reader
	Writer
	Segmenter

	io.Closer
}

type hashIndex map[string]int64

type datastore struct {
	sync.Mutex

	dir      string
	segments map[string]Segment
	current  Segment
	merging  bool

	DSOptions
}

func New(dir string, opts ...DSOption) (Datastore, error) {
	ds := &datastore{
		dir:      dir,
		segments: make(map[string]Segment),

		DSOptions: DSOptions{
			Context: context.Background(),
		},
	}

	runtime.SetFinalizer(ds, (*datastore).Close)

	err := ds.Recover()
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	ds.current = ds.NewSegment()

	return ds.WithOptions(opts...), nil
}

func (d *datastore) WithOptions(opts ...DSOption) Datastore {
	for _, opt := range opts {
		opt(&d.DSOptions)
	}

	return d
}

func (d *datastore) Options() DSOptions {
	return d.DSOptions
}

func (d *datastore) Recover() error {
	files, err := ioutil.ReadDir(d.dir)
	if err != nil {
		return err
	}

	var segments []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		segments = append(segments, file.Name())
	}

	for _, name := range segments {
		var file *os.File

		file, _, err = openFile(d.dir, name)
		if err != nil {
			return err
		}

		s := NewSegment(file)
		if err = s.Recover(); err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		d.segments[name] = s
		d.current = s
	}

	return err
}

func (d *datastore) Get(key string) ([]byte, error) {
	d.Lock()
	segments := d.segments
	d.Unlock()

	for _, s := range segments {
		val, err := s.Get(key)
		if err == nil {
			if val == nil {
				break
			}
			return val, nil
		} else if errors.Is(err, io.EOF) {
			continue
		}
	}

	return nil, ErrNotFound
}

func (d *datastore) Has(key string) (bool, error) {
	val, err := d.Get(key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (d *datastore) Put(key string, value []byte) error {
	if !d.current.IsOpen() {
		if err := d.current.Close(); err != nil {
			return err
		}

		d.current = d.NewSegment()
	}

	if err := d.current.Write(entry{key: key, value: value}); err != nil {
		return err
	}

	return nil
}

func (d *datastore) Delete(key string) error {
	return d.Put(key, nil)
}

func (d *datastore) NewSegment() Segment {
	name := strconv.FormatInt(time.Now().UnixNano(), 10)

	file, _, err := openFile(d.dir, name)
	if err != nil {
		return nil
	}

	s := NewSegment(file)

	d.Lock()
	d.segments[name] = s
	d.Unlock()

	if len(d.segments) >= 3 {
		go d.mergeSegments()
	}

	return s
}

func (d *datastore) Close() error {
	d.Lock()
	defer d.Unlock()

	// Clear the finalizer.
	runtime.SetFinalizer(d, nil)

	for _, s := range d.segments {
		if err := s.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (d *datastore) mergeSegments() {
	d.Lock()
	defer d.Unlock()

	if d.merging {
		return
	}

	d.merging = true
	defer func() { d.merging = false }()

	// TODO
}
