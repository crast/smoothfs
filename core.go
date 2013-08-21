package smoothfs

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ = io.Copy

// FS implements the hello world file system.
type SmoothFS struct {
	SrcDir    string
	CacheDir  string
	NumSlaves int
	io_queue  chan IOReq
}

func (fs *SmoothFS) Root() (fs.Node, fuse.Error) {
	log.Printf("Asked for root\n")
	return &Dir{FS: fs, RelPath: "", AbsPath: fs.SrcDir}, nil
}

func (fs *SmoothFS) Setup() {
	if fs.io_queue == nil {
		fs.io_queue = make(chan IOReq)
		for i := 0; i < fs.NumSlaves; i++ {
			go io_slave(fs, i, fs.io_queue)
		}
	}
}

func (fs *SmoothFS) Destroy() {
	close(fs.io_queue)
}

func (fs *SmoothFS) Init(req *fuse.InitRequest, resp *fuse.InitResponse, intr fs.Intr) fuse.Error {
	log.Printf("In init")
	fs.Setup()
	resp.Flags |= fuse.InitAsyncRead
	resp.MaxWrite = BLOCK_SIZE
	return nil
}

func (SmoothFS) queue() {

}

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	FS      *SmoothFS
	RelPath string
	AbsPath string
}

func (d *Dir) Attr() fuse.Attr {
	info, err := os.Stat(d.AbsPath)
	if err != nil {
		return fuse.Attr{}
	}
	return fuseAttrFromStat(info).Attr
}

func (d *Dir) Lookup(name string, intr fs.Intr) (fs.Node, fuse.Error) {
	log.Printf("In lookup\n")
	absPath := filepath.Join(d.AbsPath, name)
	relPath := filepath.Join(d.RelPath, name)
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fuse.ENOENT
	} else if info.IsDir() {
		return &Dir{FS: d.FS, RelPath: relPath, AbsPath: absPath}, nil
	} else {
		return &File{FS: d.FS, RelPath: relPath, AbsPath: absPath}, nil
	}
}

func (d *Dir) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {
	log.Printf("In readdir\n")
	fp, err := os.Open(d.AbsPath)
	if err != nil {
		log.Printf("error %s\n", err.Error())
		return nil, fuse.ENOENT
	}
	infos, rd_err := fp.Readdir(0)
	if rd_err != nil {
		log.Printf("Read error %s\n", rd_err.Error())
		return nil, fuse.EIO
	}
	dirs := make([]fuse.Dirent, 0, len(infos))
	for _, info := range infos {
		attr := fuseAttrFromStat(info)
		ent := fuse.Dirent{
			Name:  attr.Name,
			Inode: attr.Inode,
			Type:  modeDT(info.Mode()),
		}
		dirs = append(dirs, ent)
	}
	log.Printf("numdirs: %d\n", len(dirs))
	return dirs, nil
}

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

// File implements both Node and Handle for the hello file.
type File struct {
	AbsPath string
	RelPath string
	FS      *SmoothFS
	fp      *os.File
	cf      *CachedFile
}

func (f *File) Attr() fuse.Attr {
	info, err := os.Stat(f.AbsPath)
	if err != nil {
		return fuse.Attr{}
	}
	return fuseAttrFromStat(info).Attr
}

func (f *File) getFP() *os.File {
	if f.fp == nil {
		f.fp, _ = os.Open(f.AbsPath)
	}
	return f.fp
}

func (f *File) getCachedFile() *CachedFile {
	if f.cf == nil {
		f.cf = NewCachedFile(f)
	}
	return f.cf
}

func (f *File) Read(req *fuse.ReadRequest, resp *fuse.ReadResponse, intr fs.Intr) fuse.Error {
	log.Println("In File.Read")
	//fp := f.getFP()
	//buf := make([]byte, req.Size)
	reqgetter := make(chan []byte)
	cf := f.getCachedFile()
	cf.ReadRequest(req.Offset, int64(req.Size), reqgetter)
	select {
	case dbytes := <-reqgetter:
		resp.Data = dbytes
		return nil
	case <-intr:
		log.Printf("Got INTR for some reason.")
		return fuse.Errno(syscall.EINTR)
	}
	/*read_bytes, err := fp.ReadAt(buf, req.Offset)
	if err != nil && err != io.EOF {
		if err == io.EOF {
			resp.Data = nil
			req.Respond(resp)
		} else {
			fmt.Println("Error: %s", err.Error())
			return fuse.EIO
		}
	})
	fmt.Printf("About to respond: %d of %d\n", read_bytes, req.Size)
	resp.Data = buf[:read_bytes]*/
	// req.Respond(resp)
	return nil
}
