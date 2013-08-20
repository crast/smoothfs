package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"crast.us/smoothfs"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s [options] src_dir MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	io_slaves := flag.Int("io_slaves", 3, "Number of I/O slaves")
	flag.Usage = Usage
	flag.Parse()

	if flag.NArg() != 2 {
		Usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(1)

	c, err := fuse.Mount(mountpoint)
	if err != nil {
		log.Fatal(err)
	}

	fs_obj := &smoothfs.SmoothFS{
		SrcDir: flag.Arg(0),
		NumSlaves: *io_slaves,
	}

	fs.Serve(c, fs_obj)
}
