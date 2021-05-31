package datastore

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

type Segment interface {
	IsOpen() bool

	Write(key string, val []byte) error
	Get(key string) (val []byte, err error)

	Recover() error
	Snapshot() (values map[string][]byte, err error)

	io.Closer
}

type Segmenter interface {
	NewSegment() Segment
}

type readWriteSeekCloser interface {
	io.ReadWriteSeeker
	io.Closer
}

type segment struct {
	sync.Mutex

	rwsc      readWriteSeekCloser
	outOffset int64
	outSize   int64

	index    hashIndex
	entriesC chan entry
	errorsC  chan error
	closed   bool
}

func NewSegment(rwsc readWriteSeekCloser) Segment {
	s := &segment{
		rwsc: rwsc,

		index:    make(hashIndex),
		entriesC: make(chan entry),
		errorsC:  make(chan error),
	}

	go s.operateEntries()

	return s
}

func (s *segment) Recover() (err error) {
	var buf [bufSize]byte
	reader := bufio.NewReaderSize(s.rwsc, bufSize)

	for err == nil {
		var (
			header, data []byte
			n            int
		)

		header, err = reader.Peek(bufSize)
		if err == io.EOF {
			if len(header) == 0 {
				return err
			}
		} else if err != nil {
			return err
		}

		size := binary.LittleEndian.Uint32(header)
		if size < bufSize {
			data = buf[:size]
		} else {
			data = make([]byte, size)
		}

		n, err = reader.Read(data)
		if err == nil {
			if n != int(size) {
				return fmt.Errorf("corrupted file")
			}

			var e entry

			e.Decode(data)

			s.index[e.key] = s.outOffset
			s.outOffset += int64(n)
		}
	}

	return err
}

func (s *segment) IsOpen() bool {
	return s.outOffset < bufSize
}

func (s *segment) Write(key string, val []byte) error {
	if s.closed {
		return ErrSegmentClosed
	}

	s.entriesC <- entry{key: key, value: val}

	return <-s.errorsC
}

func (s *segment) Get(key string) ([]byte, error) {
	s.Lock()
	position, ok := s.index[key]
	s.Unlock()

	if !ok {
		return nil, ErrNotFound
	}

	if _, err := s.rwsc.Seek(position, 0); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(s.rwsc)

	value, err := readValue(reader)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *segment) Snapshot() (map[string][]byte, error) {
	all := make(map[string][]byte)

	for key := range s.index {
		val, err := s.Get(key)
		if err != nil {
			return nil, err
		}

		all[key] = val
	}

	return all, nil
}

func (s *segment) Close() error {
	s.Lock()
	defer s.Unlock()

	close(s.entriesC)

	s.entriesC = nil
	s.closed = true

	return s.rwsc.Close()
}

func (s *segment) operateEntries() {
	for e := range s.entriesC {
		s.Lock()

		n, err := s.rwsc.Write(e.Encode())
		if err != nil {
			s.errorsC <- err
			continue
		}

		s.index[e.key] = s.outOffset
		s.outOffset += int64(n)

		s.errorsC <- nil
		s.Unlock()
	}
}
