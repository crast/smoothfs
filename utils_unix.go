
// +build linux darwin

package smoothfs

import "syscall"
import "os"
import "bazil.org/fuse"

func fuseAttrFromStat_unix(info *os.FileInfo, attr *fuse.Attr, raw_stat *syscall.Stat_t) {
	attr.Inode = raw_stat.Ino
	attr.Nlink = uint32(raw_stat.Nlink)
	attr.Uid = raw_stat.Uid
	attr.Gid = raw_stat.Gid
	attr.Rdev = uint32(raw_stat.Rdev)
	attr.Blocks = uint64(raw_stat.Blocks)
}
