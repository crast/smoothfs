package smoothfs

import (
	"bazil.org/fuse"
)

const BLOCK_SIZE = 65536

type BlockNum int

func (b BlockNum) Offset() int64 {
	return int64(b) * BLOCK_SIZE
}

type IOReq struct {
	Node File
	BlockNum BlockNum
	Result []byte
}

type fileattrs struct {
	fuse.Attr
	Name string
}
