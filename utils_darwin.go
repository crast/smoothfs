
// +build darwin

package smoothfs

import "syscall"
import "os"
import "bazil.org/fuse"
import "time"

/* &syscall.Stat_t{Dev:16777218, Mode:0x41ed, Nlink:0x6, Ino:0x16013e, 
		Uid:0x1f5, Gid:0x14, Rdev:0, Pad_cgo_0:[4]uint8{0x0, 0x0, 0x0, 0x0}, 
		Atimespec:syscall.Timespec{Sec:1375337156, Nsec:0},
		 Mtimespec:syscall.Timespec{Sec:1369414194, Nsec:0}, 
		 Ctimespec:syscall.Timespec{Sec:1369414194, Nsec:0}, 
		 Birthtimespec:syscall.Timespec{Sec:1369414179, Nsec:0}, 
		 Size:204, Blocks:0, Blksize:4096, Flags:0x0, 
		 Gen:0x0, Lspare:0, Qspare:[2]int64{0, 0}}*/
    
func fuseAttrFromStat_inner(info *os.FileInfo, attr *fuse.Attr, raw_stat *syscall.Stat_t) {
	fuseAttrFromStat_unix(info, attr, raw_stat)
	attr.Ctime = time.Unix(raw_stat.Ctimespec.Unix())
	attr.Crtime = time.Unix(raw_stat.Birthtimespec.Unix())
	attr.Flags = raw_stat.Flags
}
