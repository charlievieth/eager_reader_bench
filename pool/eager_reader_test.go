package pool

import (
	"io"
	"runtime"
	"testing"
)

const (
	kB int = 1 << (10 * (iota + 1))
	mB
	gB
)

const (
	FragSize = mB * 10
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

func (r *ReadCloser) Read(p []byte) (int, error) {
	if r.off >= len(r.buf) {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n := copy(p, r.buf[r.off:])
	r.off += n
	return n, nil
}

func (r *ReadCloser) Reset() {
	r.off = 0
}

func (r *ReadCloser) Close() error {
	return nil
}

func BenchmarkEagerReader(b *testing.B) {
	b.ReportAllocs()
	rc := &ReadCloser{
		buf: make([]byte, FragSize),
	}
	w := &NoopWriter{}
	count := BodySize / FragSize
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := newPool(2, int64(FragSize))
		for j := 0; j < count; j++ {
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
}
