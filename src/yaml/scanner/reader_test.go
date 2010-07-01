package scanner

import (
	"bytes"
	"io"
	"testing"
	"testing/iotest"
)

func TestCacheFull(t *testing.T) {
	const data = "Spam, Eggs, Bacon, Spam, and Spam"
	r := newReader(bytes.NewBufferString(data))
	if r.Len() != 0 {
		t.Fatal("Reader starts off with a cache")
	}
	// CacheFull one byte
	if err := r.CacheFull(1); err != nil {
		t.Error("CacheFull error:", err)
	}
	if r.Len() != 1 {
		t.Errorf("Reader asked to cache 1 byte, have %d bytes", r.Len())
	}
	// CacheFull one more byte
	if err := r.CacheFull(2); err != nil {
		t.Error("CacheFull error:", err)
	}
	cacheSize := r.Len()
	if r.Len() != 2 {
		t.Errorf("Reader asked to cache 2 bytes, have %d bytes", r.Len())
	}
	// This shouldn't cache anything more.
	if err := r.CacheFull(1); err != nil {
		t.Error("CacheFull error:", err)
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
		t.Error("CacheFulld read corrupted")
	}
	if err != nil {
		t.Error(err)
	}
}

func TestHalfCacheFull(t *testing.T) {
	r := newReader(bytes.NewBufferString("Hi"))
	err := r.CacheFull(4)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("CacheFull should give unexpected EOF error, instead got %v", err)
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
	r.CacheFull(1)
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
	r.CacheFull(cacheSize)
	if !r.Check(0, data[0:cacheSize]) {
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

func TestReadBreak(t *testing.T) {
	r := newReader(bytes.NewBufferString("a\rb\nc\r\nd\n"))
	// Non-break ReadBreak
	{
		bytes, err := r.ReadBreak()
		if err != nil {
			t.Error("First ReadBreak error:", err)
		} else if len(bytes) != 0 {
			t.Error("First ReadBreak yielded non-break characters")
		}
		if c, _ := r.ReadByte(); c != 'a' {
			t.Fatal("SkipBreak skips non-break characters")
		}
	}
	// Break check function
	check := func(name, br string, nextChar byte) {
		bytes, err := r.ReadBreak()
		if err != nil {
			t.Error(name, "ReadBreak error:", err)
		} else if string(bytes) != br {
			t.Errorf("%s ReadBreak gave wrong break: %#v", name, string(bytes))
		}
		if nextChar != 0 {
			if c, _ := r.ReadByte(); c != nextChar {
				t.Errorf("Didn't skip %s", name)
				r.ReadByte()
			}
		}
	}
	check("CR", "\r", 'b')
	check("LF", "\n", 'c')
	check("CRLF", "\r\n", 'd')
	check("Final LF", "\n", 0)
	// Test final CR
	r = newReader(bytes.NewBufferString("a\r"))
	r.ReadByte()
	check("Final CR", "\r", 0)
}
