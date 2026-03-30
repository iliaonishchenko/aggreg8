// Package router содержит HTTP-middleware для сжатия и подписи запросов/ответов.
package router

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

var gzipReaderPool = sync.Pool{}

// compressWriter
type compressWriter struct {
	w        http.ResponseWriter
	zw       *gzip.Writer
	compress bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.compress {
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Close() error {
	if c.compress {
		err := c.zw.Close()
		gzipWriterPool.Put(c.zw)
		return err
	}
	gzipWriterPool.Put(c.zw)
	return nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	contentType := c.w.Header().Get("Content-Type")
	c.compress = statusCode < 300 &&
		(strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html"))
	if c.compress {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// compressReader

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	if v := gzipReaderPool.Get(); v != nil {
		zr := v.(*gzip.Reader)
		if err := zr.Reset(r); err != nil {
			return nil, err
		}
		return &compressReader{r: r, zr: zr}, nil
	}
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r: r, zr: zr}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	err := c.zr.Close()
	gzipReaderPool.Put(c.zr)
	return err
}

// WithCompression — middleware, обеспечивающий gzip-сжатие запросов и ответов.
func WithCompression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	})
}
