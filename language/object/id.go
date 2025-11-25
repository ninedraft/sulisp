package object

import "sync/atomic"

var objectIDs atomic.Uint64

func makeObjectID() uint64 {
	return objectIDs.Add(1)
}
