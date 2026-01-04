package bouteitech

import (
	"net"
	"net/http"
)

// flushWriter は net.Conn を http.ResponseWriter として振る舞わせるためのアダプターです
type flushWriter struct {
	net.Conn
}

func (fw *flushWriter) Header() http.Header {
	return http.Header{}
}

func (fw *flushWriter) Write(p []byte) (int, error) {
	return fw.Conn.Write(p)
}

func (fw *flushWriter) WriteHeader(_ int) {
	// http.Error が内部でこれを呼び出しますが、
	// CONNECTメソッドの前段階では生で書き込むため、ここでは何もしなくても動作します。
}
