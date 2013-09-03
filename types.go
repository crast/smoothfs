package smoothfs

import (
	"bazil.org/fuse"
)

const BLOCK_SIZE = 65536

// a BlockNum is a numerical address of one block inside a (Cached)File.
type BlockNum int

// The Offset of this BlockNum, in bytes from beginning of a file.
func (b BlockNum) Offset() int64 {
	return int64(b) * BLOCK_SIZE
}

type WorkEntry interface {
	Process()
	Responder() chan WorkEntry
}

type fileattrs struct {
	fuse.Attr
	Name string
}
