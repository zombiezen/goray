package scanner

import (
	"bytes"
	"io"
	"testing"
	"testing/iotest"
)

func TestCache(t *testing.T) {
	const data = "Spam, Eggs, Bacon, Spam, and Spam"
	r := newReader(bytes.NewBufferString(data))
	if r.Len() != 0 {
		t.Fatal("Reader starts off with a cache")
	}
	// Cache one byte
	if err := r.Cache(1); err != nil {
		t.Error("Cache error:", err)
	}
	if r.Len() != 1 {
		t.Errorf("Reader asked to cache 1 byte, have %d bytes", r.Len())
	}
	// Cache one more byte
	if err := r.Cache(2); err != nil {
		t.Error("Cache error:", err)
	}
	cacheSize := r.Len()
	if r.Len() != 2 {
		t.Errorf("Reader asked to cache 2 bytes, have %d bytes", r.Len())
	}
	// This shouldn't cache anything more.
	if err := r.Cache(1); err != nil {
		t.Error("Cache error:", err)
	}
	if r.Len() != cacheSize {
		t.Error("Redundant cache changed buffer")
	}
	if string(r.Bytes()) != data[0:cacheSize] {
		t.Error("Buffer is reading the wrong data")
	}
	// Use buffer to read
	result := make([]byte, cacheSize+1)
	nRead, err := r.Read(result)
	if nRead != cacheSize {
		t.Errorf("Read from cache gave %d bytes, expected %d", nRead, cacheSize)
	}
	if string(result[0:nRead]) != data[0:nRead] {
		t.Error("Cached read corrupted")
	}
	if err != nil {
		t.Error(err)
	}
}

func TestHalfCache(t *testing.T) {
	r := newReader(bytes.NewBufferString("Hi"))
	err := r.Cache(4)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Cache should give unexpected EOF error, instead got %v", err)
	}
}

func TestInitialPos(t *testing.T) {
	r := newReader(bytes.NewBufferString(""))
	if r.Pos.Index != 0 {
		t.Error("Initial index is wrong")
	}
	if r.Pos.Column != 1 {
		t.Error("Initial column is wrong")
	}
	if r.Pos.Line != 1 {
		t.Error("Initial line number is wrong")
	}
}

func TestPos(t *testing.T) {
	const data = "Hello\nGoodbye"
	r := newReader(bytes.NewBufferString(data))
	r.Cache(1)
	if r.Pos.Index != 0 || r.Pos.Column != 1 || r.Pos.Line != 1 {
		t.Error("Caching moves position")
	}
	if r.Next(1); r.Pos.Index != 1 || r.Pos.Column != 2 || r.Pos.Line != 1 {
		t.Error("Next doesn't move position")
	}
	if r.ReadByte(); r.Pos.Index != 2 || r.Pos.Column != 3 || r.Pos.Line != 1 {
		t.Error("ReadByte doesn't move position")
	}
	myData := make([]byte, 4)
	io.ReadFull(r, myData)
	if r.Pos.Index != 6 {
		t.Error("Read doesn't update position")
	}
	if r.Pos.Column == 7 || r.Pos.Line == 1 {
		t.Error("Read doesn't update line number properly")
	}
}

func TestOneByteReader(t *testing.T) {
	const (
		data      = "Hello, World!"
		cacheSize = 5
	)

	b := bytes.NewBufferString(data)
	r := newReader(iotest.OneByteReader(b))
	r.Cache(cacheSize)
	if !r.Check(data[0:cacheSize]) {
		t.Error("Caching failed")
	}
	result := make([]byte, len(data))
	nRead, err := io.ReadFull(r, result)
	if nRead != len(data) {
		t.Error("Full read failed")
	}
	if string(result) != data {
		t.Errorf("Read corrupted. (wanted %#v, got %#v)", data, string(result))
	}
	if err != nil {
		t.Error(err)
	}
}
