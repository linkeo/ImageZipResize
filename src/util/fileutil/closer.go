package fileutil

import "io"

func NewReadCloser(r io.Reader) io.ReadCloser {
	rc, ok := r.(io.ReadCloser)
	if ok {
		return rc
	}
	return &readCloser{r: r}
}

func NewWriteCloser(w io.Writer) io.WriteCloser {
	wc, ok := w.(io.WriteCloser)
	if ok {
		return wc
	}
	return &writeCloser{w: w}
}

type readCloser struct {
	r io.Reader
}

func (r readCloser) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r readCloser) Close() error {
	c, ok := r.r.(io.Closer)
	if !ok {
		return nil
	}
	return c.Close()
}

type writeCloser struct {
	w io.Writer
}

func (w writeCloser) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func (w writeCloser) Close() error {
	c, ok := w.w.(io.Closer)
	if !ok {
		return nil
	}
	return c.Close()
}
