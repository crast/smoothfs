package smoothfs

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

// struct Block represents one block in a CachedFile.
type Block struct {
	OnDisk bool
	Loaded bool
	bytes  []byte
}

// A CachedFile connects a backing file with a local cache file and memory cache.
// This allows files to be read in more sensible chunks than applications
// often request and avoids a lot of round trips in the case of slower
// filesystems such as network and other remote filesystems.
// This is the primary mover and shaker of the SmoothFS ecosystem.
type CachedFile struct {
	*SmoothFS
	SrcFilePath   string // The absolute path of the file we're caching
	CacheFilePath string // The absolute path of the cache block file.
	fp            *os.File
	blocks        map[BlockNum]*Block
	dirty_blocks  []BlockNum
}

// ReadRequest begins a new read request which it will respond to on responder.
func (cf *CachedFile) ReadRequest(offset int64, length int, responder chan []byte) {
	go (func() {
		data := cf.Read(offset, length)
		if data != nil {
			responder <- data
		}
	})()
}

// Read performs the actual mechanism of reading, and returns the bytes read or nil.
// offset is always the offset from the beginning of a file and must be a positive number.
// length is the amount of bytes to read and must also be a positive number.
func (cf *CachedFile) Read(offset int64, length int) []byte {
	start_block := loc_in_block(offset)
	end_block := loc_in_block(offset + int64(length) - 1)
	var to_retrieve []BlockNum = nil
	for i := start_block; i <= end_block; i++ {
		if blk := cf.blocks[i]; blk == nil || !blk.Loaded {
			to_retrieve = append(to_retrieve, i)
		}
	}
	if len(to_retrieve) != 0 {
		if !cf.RetrieveBlocks(to_retrieve) {
			return nil
		}
	}

	offsetA := offset - start_block.Offset()
	offsetB := offset + int64(length) - (end_block.Offset())
	log.Printf("Offset: %d, length:%d, start_block:%d, end_block:%d, offsetA:%d, offsetB:%d",
		offset, length, start_block, end_block, offsetA, offsetB)
	if start_block == end_block {
		bbytes := cf.blocks[start_block].bytes
		num_bytes := len(bbytes)
		if int64(num_bytes) >= offsetB {
			return bbytes[offsetA:offsetB]
		} else if int64(num_bytes) >= offsetA {
			return bbytes[offsetA:]
		} else {
			log.Printf("OffsetA %d and offsetB %d not worky blockum %d %#v", offsetA, offsetB, start_block, bbytes)
			log.Fatal("Bye.")
			return nil
		}
	} else {
		buffer := bytes.NewBuffer(cf.blocks[start_block].bytes[offsetA:])
		for i := start_block + 1; i < end_block; i++ {
			buffer.Write(cf.blocks[i].bytes)
		}
		buffer.Write(cf.blocks[end_block].bytes[:offsetB])
		return buffer.Bytes()
	}
	return nil
}

// RetrieveBlocks gets blocks from the disk representation of this file and wait for results.
func (cf *CachedFile) RetrieveBlocks(blocks []BlockNum) bool {
	signaler := make(chan WorkEntry)
	io_queue := cf.SmoothFS.io_queue
	for _, blocknum := range blocks {
		io_queue <- &CacheIOReq{
			CachedFile: cf,
			BlockNum:   blocknum,
			responder:  signaler,
		}
	}
	for i := len(blocks); i > 0; i-- {
		f := <-signaler
		log.Printf("Got IOReq response for block %d", f.(*CacheIOReq).BlockNum)
	}
	return true
}

// internalRead performs the actual reading part of handling an IO request.
func (cf *CachedFile) internalRead(req CacheIOReq) {
	rbytes := cf.reallyRead(req.BlockNum.Offset(), BLOCK_SIZE)
	if rbytes == nil {
		return
	}
	cf.blocks[req.BlockNum] = &Block{
		Loaded: true,
		bytes:  rbytes,
	}
	cf.dirty_blocks = append(cf.dirty_blocks, req.BlockNum)
}

// reallyRead is the low level read function, handles at the byte level.
func (cf *CachedFile) reallyRead(offset int64, length int) []byte {
	fp := cf.getFile()
	buf := make([]byte, length)
	n, err := fp.ReadAt(buf, offset)
	if err != nil {
		if err == io.EOF {
			if n == 0 {
				return nil
			}
		} else {
			log.Fatal(err)
		}
	}
	if n < length {
		return buf[:n]
	} else {
		return buf
	}
}

// getFile gets the internal file pointer of this CachedFile.
func (cf *CachedFile) getFile() *os.File {
	if cf.fp == nil {
		fp, err := os.Open(cf.SrcFilePath)
		if err != nil {
			log.Fatal(err)
		}
		cf.fp = fp
	}
	return cf.fp
}

// NewCachedFile creates a CachedFile entity from a SmoothFS file.
func NewCachedFile(f *File) *CachedFile {
	return &CachedFile{
		SmoothFS:      f.FS,
		SrcFilePath:   f.AbsPath,
		CacheFilePath: filepath.Join(f.FS.CacheDir, f.RelPath),
		blocks:        make(map[BlockNum]*Block),
	}

}


// An IOReq is a block read that is handled on one of the background IO slaves.
type CacheIOReq struct {
	*CachedFile            // The CachedFile which is doing this read request.
	BlockNum    BlockNum   // The block number to read.
	responder   chan WorkEntry // A channel we are expected to respond on.
}

func (req *CacheIOReq) Process() {
	req.internalRead(*req)
	if len(req.dirty_blocks) > 2 {
		req.SmoothFS.clean_queue <- &CacheCleanReq{req.CachedFile}
	}
}

func (req *CacheIOReq) Responder() chan WorkEntry {
	return req.responder
}


type CacheCleanReq struct {
	*CachedFile
}

func (req *CacheCleanReq) Process() {
	db := req.CachedFile.dirty_blocks
	l := len(db)
	if l == 0 {
		return
	} else if l == 1 {
		req.CachedFile.dirty_blocks = nil
	} else {
		req.CachedFile.dirty_blocks = db[1:]
	}
	blockNum := db[0]
	block := req.blocks[blockNum]
	if block.Loaded {
		log.Printf("Going to unload block %d")
	}
}

func (req *CacheCleanReq) Responder() chan WorkEntry {
	return nil
}
