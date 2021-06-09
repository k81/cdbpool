package cdbpool

import (
	"sync"

	"github.com/std0d9k81/sqlparser"
)

var (
	gBufferPool *sync.Pool
)

func GetBuffer() *sqlparser.TrackedBuffer {
	return gBufferPool.Get().(*sqlparser.TrackedBuffer)
}

func PutBuffer(buf *sqlparser.TrackedBuffer) {
	buf.Reset()
	gBufferPool.Put(buf)
}

func init() {
	gBufferPool = &sync.Pool{
		New: func() interface{} {
			return sqlparser.NewTrackedBuffer(nil)
		},
	}
}
