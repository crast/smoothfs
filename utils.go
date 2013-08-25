package smoothfs

import (
	"bazil.org/fuse"
	"fmt"
	"os"
	"syscall"
)

func fuseAttrFromStat(info os.FileInfo) fileattrs {
	attr := fuse.Attr{
		Size:  uint64(info.Size()),
		Mode:  info.Mode(),
		Mtime: info.ModTime(),
	}
	bits, c_ok := info.Sys().(*syscall.Stat_t)
	if c_ok {
		fuseAttrFromStat_inner(&info, &attr, bits)
	} else {
		fmt.Printf("%#v", info.Sys())
	}
	attrs := fileattrs{
		Name: info.Name(),
		Attr: attr,
	}
	return attrs
}

// loc_in_block takes a byte offset and turns it into a BlockNum.
func loc_in_block(loc int64) BlockNum {
	return BlockNum(loc / BLOCK_SIZE)
}

// modeDT takes an os.FileMode and gives the appropriate DirEntType.
func modeDT(mode os.FileMode) fuse.DirentType {
	if mode.IsDir() {
		return fuse.DT_Dir
	} else if mode.IsRegular() {
		return fuse.DT_File
	} else if mode&os.ModeDevice != 0 {
		return fuse.DT_Block
	} else {
		return fuse.DT_Unknown
	}
}
