package main

import (
	"io"
	"runtime"
)

const (
	kB int = 1 << (10 * (iota + 1))
	mB
	gB
)

const (
	FragSize = mB * 20
	BodySize = mB * 1000
)

var _ = runtime.GOMAXPROCS(1)

type NoopWriter struct{}

func (n *NoopWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

type ReadCloser struct {
	off int
	buf []byte
}

func (r *ReadCloser) Read(p []byte) (n int, err error) {
	if r.off >= len(r.buf) {
		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}
	n = copy(p, r.buf[r.off:])
	r.off += n
	return n, nil
}

func (r *ReadCloser) Reset() {
	r.off = 0
}

func (r *ReadCloser) Close() error {
	return nil
}

func main() {
	rc := &ReadCloser{
		buf: make([]byte, FragSize),
	}
	w := &NoopWriter{}
	c := BodySize / FragSize
	p := newPool(1, int64(FragSize))
	for i := 0; i < c; i++ {
		e := newEagerReader(rc, int64(FragSize), p)
		for {
			if _, err := e.writeOnce(w); err != nil {
				e.free()
				break
			}
		}
		rc.Reset()
	}
}

/*
func BenchmarkEagerReader(b *testing.B) {
	rc := &ReadCloser{
		buf: make([]byte, FragSize),
	}
	w := &NoopWriter{}
	c := BodySize / FragSize
	b.ResetTimer()
	for i := 0; i < c; i++ {
		e := newEagerReader(rc, int64(FragSize))
		for {
			if _, err := e.writeOnce(w); err != nil {
				break
			}
		}
		rc.Reset()
	}
}
*/
