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

	Write(e entry) error
	Get(key string) (val []byte, err error)

	Recover() error

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
	closed   bool
}

func NewSegment(rwsc readWriteSeekCloser) Segment {
	s := &segment{
		rwsc: rwsc,

		index:    make(hashIndex),
		entriesC: make(chan entry),
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

func (s *segment) Write(e entry) error {
	if s.closed {
		return ErrSegmentClosed
	}

	s.entriesC <- e

	return nil
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

func (s *segment) Close() error {
	s.Lock()
	defer s.Unlock()

	close(s.entriesC)

	s.entriesC = nil
	s.closed = true

	return nil
}

func (s *segment) operateEntries() {
	for e := range s.entriesC {
		s.Lock()

		n, err := s.rwsc.Write(e.Encode())
		if err != nil {
			continue
		}

		s.index[e.key] = s.outOffset
		s.outOffset += int64(n)

		s.Unlock()
	}

	return
}
