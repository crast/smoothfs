package smoothfs

import (
	"os"
	"path/filepath"
	"fmt"
	"syscall"
	"time"
	"io"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// FS implements the hello world file system.
type SmoothFS struct{
	SrcDir string
}

func (fs *SmoothFS) Root() (fs.Node, fuse.Error) {
	fmt.Printf("Asked for root\n")
	return &Dir{FS: fs, RelPath: "", AbsPath: fs.SrcDir}, nil
}


func (SmoothFS) queue() {

}

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	FS *SmoothFS
	RelPath string
	AbsPath string
}

func (d *Dir) Attr() fuse.Attr {
	info, err := os.Stat(d.AbsPath)
	if (err != nil) {
		return fuse.Attr{}
	}
	return fuseAttrFromStat(info).Attr
}

func (d *Dir) Lookup(name string, intr fs.Intr) (fs.Node, fuse.Error) {
	fmt.Printf("In lookup\n")
	absPath := filepath.Join(d.AbsPath, name)
	relPath := filepath.Join(d.RelPath, name)
	info, err := os.Stat(absPath)
	if (err != nil) {
		return nil, fuse.ENOENT
	} else if info.IsDir() {
		return &Dir{FS: d.FS, RelPath: relPath, AbsPath: absPath}, nil
	} else {
		return &File{RelPath: relPath, AbsPath: absPath}, nil
	}
}

func (d *Dir) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {
	fmt.Printf("In readdir\n")
	fp, err := os.Open(d.AbsPath)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
		return nil, fuse.ENOENT
	}
	infos, rd_err := fp.Readdir(0)
	if rd_err != nil {
		fmt.Printf("Read error %s\n", rd_err.Error())
		return nil, fuse.EIO
	}
	dirs := make([]fuse.Dirent, 0, len(infos))
	for _, info := range infos {
		attr := fuseAttrFromStat(info)
		ent := fuse.Dirent{
			Name: attr.Name,
			Inode: attr.Inode,
			Type: modeDT(info.Mode()),
		}
		dirs = append(dirs, ent)
	}
	fmt.Printf("numdirs: %d\n", len(dirs))
	return dirs, nil
}

func modeDT(mode os.FileMode) fuse.DirentType {
	if mode.IsDir() {
		return fuse.DT_Dir
	} else if mode.IsRegular() {
		return fuse.DT_File
	} else if mode & os.ModeDevice != 0 {
		return fuse.DT_Block
	} else {
		return fuse.DT_Unknown
	}
}

func fuseAttrFromStat(info os.FileInfo) (fileattrs) {
	/* &syscall.Stat_t{Dev:16777218, Mode:0x41ed, Nlink:0x6, Ino:0x16013e, 
		Uid:0x1f5, Gid:0x14, Rdev:0, Pad_cgo_0:[4]uint8{0x0, 0x0, 0x0, 0x0}, 
		Atimespec:syscall.Timespec{Sec:1375337156, Nsec:0},
		 Mtimespec:syscall.Timespec{Sec:1369414194, Nsec:0}, 
		 Ctimespec:syscall.Timespec{Sec:1369414194, Nsec:0}, 
		 Birthtimespec:syscall.Timespec{Sec:1369414179, Nsec:0}, 
		 Size:204, Blocks:0, Blksize:4096, Flags:0x0, 
		 Gen:0x0, Lspare:0, Qspare:[2]int64{0, 0}}*/
    
	attr := fuse.Attr{
		Size: uint64(info.Size()),
		Mode: info.Mode(),
		Mtime: info.ModTime(),
	}
	bits, c_ok := info.Sys().(*syscall.Stat_t)
	if c_ok {
		attr.Inode = bits.Ino
		attr.Ctime = time.Unix(bits.Ctimespec.Sec, 0)
		attr.Nlink = uint32(bits.Nlink)
		attr.Uid = bits.Uid
		attr.Gid = bits.Gid
	} else {
		fmt.Printf("%#v", info.Sys())
	}
	attrs := fileattrs{
    	Name: info.Name(),
    	Attr: attr,
    }
	return attrs

}

type fileattrs struct {
	fuse.Attr
	Name string
}


// File implements both Node and Handle for the hello file.
type File struct {
	AbsPath string
	RelPath string
	fp *os.File
}

func (f *File) Attr() fuse.Attr {
	info, err := os.Stat(f.AbsPath)
	if (err != nil) {
		return fuse.Attr{}
	}
	return fuseAttrFromStat(info).Attr}

func (f *File) getFP() *os.File {
	if (f.fp == nil) {
		f.fp, _ = os.Open(f.AbsPath)
	}
	return f.fp
}

func (f *File) Read(req *fuse.ReadRequest, resp *fuse.ReadResponse, intr fs.Intr) fuse.Error {
	fmt.Println("In File.Read")
	fp := f.getFP()
	buf := make([]byte, req.Size)
	read_bytes, err := fp.ReadAt(buf, req.Offset)
	if err != nil && err != io.EOF {
		if err == io.EOF {
			resp.Data = nil
			req.Respond(resp)
		} else {
			fmt.Println("Error: %s", err.Error())
			return fuse.EIO
		}
	}
	fmt.Printf("About to respond: %d of %d\n", read_bytes, req.Size)
	resp.Data = buf[:read_bytes]
	req.Respond(resp)
	return nil
}
