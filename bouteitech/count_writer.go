package bouteitech

import (
	"sync/atomic"
)

// countWriter はポインタ経由で外部の変数を直接更新します
type countWriter struct {
	count *int64 // 合計バイト数へのポインタ
}

// Write は受け取ったバイト数を、ポインタ先の変数にアトミックに加算します
func (cw countWriter) Write(p []byte) (int, error) {
	n := len(p)
	if cw.count != nil {
		atomic.AddInt64(cw.count, int64(n))
	}
	return n, nil
}
