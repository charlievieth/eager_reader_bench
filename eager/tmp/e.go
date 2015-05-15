package eager

import (
	"fmt"
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
	count := BodySize / FragSize
	var c int64
	for i := 0; i < count; i++ {
		e := newEagerReader(rc, int64(FragSize))
		for {
			n, err := e.writeOnce(w)
			c += n
			if err != nil {
				break
			}
		}
		rc.Reset()
	}
	fmt.Println(c, c/int64(mB))
}

/*
var err error
		for err == nil {
			n, err = e.writeOnce(w)
			c += n
		}
*/
