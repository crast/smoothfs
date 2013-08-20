package smoothfs

import (
	"syscall"
	"os"
	"bazil.org/fuse"
	"fmt"
)

func fuseAttrFromStat(info os.FileInfo) (fileattrs) {
	attr := fuse.Attr{
		Size: uint64(info.Size()),
		Mode: info.Mode(),
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
