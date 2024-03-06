package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
)

type decompressReader struct {
	*gzip.Reader
	io.Closer
}

func (gz decompressReader) Close() error {
	return gz.Closer.Close()
}

func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			r.Body = decompressReader{gz, r.Body}
		}
		next.ServeHTTP(w, r)
	})
}
