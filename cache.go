package smoothfs

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Block struct {
	loaded bool
	bytes  []byte
}

type CachedFile struct {
	*SmoothFS
	blocks         map[BlockNum]*Block
	cachefile_path string
	srcfile_path   string
	fp             *os.File
	last_loc       int64
}

func (cf *CachedFile) ReadRequest(offset int64, length int64, responder chan []byte) {
	go (func() {
		data := cf.Read(offset, length)
		if data != nil {
			responder <- data
		}
	})()
}

func (cf *CachedFile) Read(offset int64, length int64) []byte {
	start_block := loc_in_block(offset)
	end_block := loc_in_block(offset + length - 1)
	var to_retrieve []BlockNum = nil
	for i := start_block; i <= end_block; i++ {
		if cf.blocks[i] == nil {
			to_retrieve = append(to_retrieve, i)
		}
	}
	if len(to_retrieve) != 0 {
		if !cf.RetrieveBlocks(to_retrieve) {
			return nil
		}
	}

	offsetA := offset - start_block.Offset()
	offsetB := offset + length - (end_block.Offset())
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

func (cf *CachedFile) RetrieveBlocks(blocks []BlockNum) bool {
	signaler := make(chan IOReq)
	for _, blocknum := range blocks {
		cf.SmoothFS.io_queue <- IOReq{
			CachedFile: cf,
			BlockNum:   blocknum,
			Responder:  signaler,
		}
	}
	for i := len(blocks); i > 0; i-- {
		f := <-signaler
		log.Printf("Got IOReq response for block %d", f.BlockNum)
	}
	return true
}

func (cf *CachedFile) internalRead(req IOReq) {
	rbytes := cf.reallyRead(req.BlockNum.Offset(), BLOCK_SIZE)
	if rbytes == nil {
		return
	}
	cf.blocks[req.BlockNum] = &Block{
		loaded: true,
		bytes:  rbytes,
	}
}

func (cf *CachedFile) reallyRead(offset int64, length int) []byte {
	fp := cf.getFile()
	if cf.last_loc != offset {
		fp.Seek(offset, 0)
	}
	buf := make([]byte, length)
	n, err := fp.Read(buf)
	if err != nil {
		if err == io.EOF {
			if n == 0 {
				return nil
			}
		}
		log.Fatal(err)
	}
	cf.last_loc = offset + int64(n)
	return buf[:n]

}

func (cf *CachedFile) getFile() *os.File {
	if cf.fp == nil {
		fp, err := os.Open(cf.srcfile_path)
		if err != nil {
			log.Fatal(err)
		}
		cf.fp = fp
	}
	return cf.fp
}

func loc_in_block(loc int64) BlockNum {
	return BlockNum(loc / BLOCK_SIZE)
}

func NewCachedFile(f *File) *CachedFile {
	return &CachedFile{
		SmoothFS:       f.FS,
		srcfile_path:   f.AbsPath,
		cachefile_path: filepath.Join(f.FS.CacheDir, f.RelPath),
		blocks:         make(map[BlockNum]*Block),
	}

}
