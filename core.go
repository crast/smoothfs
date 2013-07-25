package smoothfs

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// FS implements the hello world file system.
type SmoothFS struct{
	SrcDir string
}

func (SmoothFS) Root() (fs.Node, fuse.Error) {
	return Dir{}, nil
}

func (SmoothFS) queue() {

}

// Dir implements both Node and Handle for the root directory.
type Dir struct{}

func (Dir) Attr() fuse.Attr {
	return fuse.Attr{Inode: 1, Mode: os.ModeDir | 0555}
}

func (Dir) Lookup(name string, intr fs.Intr) (fs.Node, fuse.Error) {
	if name == "hello" {
		return File{}, nil
	}
	return nil, fuse.ENOENT
}

var dirDirs = []fuse.Dirent{
	{Inode: 2, Name: "hello", Type: fuse.DT_File},
}

func (Dir) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {
	return dirDirs, nil
}

// File implements both Node and Handle for the hello file.
type File struct{}

func (File) Attr() fuse.Attr {
	return fuse.Attr{Mode: 0444}
}

func (File) ReadAll(intr fs.Intr) ([]byte, fuse.Error) {
	return []byte("hello, world\n"), nil
}