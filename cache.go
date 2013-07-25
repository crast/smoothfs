package smoothfs

import (
	"bytes"
	"os"
	"log"
)


type Block struct {
	loaded bool
	bytes []byte
}

type CachedFile struct {
	*SmoothFS
	blocks map[BlockNum]*Block
	cachefile_path string
	srcfile_path string
	fp *os.File
	last_loc int64
}

func (cf *CachedFile) Read(offset int64, length int64) []byte {
	start_block := loc_in_block(offset)
	end_block := loc_in_block(offset + length - 1)
	var to_retrieve []BlockNum = nil
	for i:= start_block; i <= end_block; i++ {
		if cf.blocks[i] == nil {
			to_retrieve = append(to_retrieve, i)
		}
	}
	if len(to_retrieve) != 0 {
		cf.RetrieveBlocks(to_retrieve)
	}

	offsetA := offset - start_block.Offset()
	offsetB := offset + length - ((end_block - 1).Offset())
	if start_block == end_block {
		return cf.blocks[start_block].bytes[offsetA:offsetB]
	} else {
		buffer := bytes.NewBuffer(cf.blocks[start_block].bytes[offsetA:])
		for i:= start_block + 1; i < end_block; i++ {
			buffer.Write(cf.blocks[i].bytes)
		}
		buffer.Write(cf.blocks[end_block].bytes[:offsetB])
	}
	return nil
}

func (cf *CachedFile) RetrieveBlocks(blocks []BlockNum) {
	for _, blocknum := range blocks {
		cf.blocks[blocknum] = &Block{
			loaded: true,
			bytes: cf.reallyRead(blocknum.Offset(), BLOCK_SIZE),
		}
	}
}

func (cf *CachedFile) reallyRead(offset int64, length int) []byte {
	fp := cf.getFile()
	if cf.last_loc != offset {
		fp.Seek(offset, 0)
	}
	buf := make([]byte, length)
	n, err := fp.Read(buf)
	if (err != nil) {
		log.Fatal(err)
	}
	cf.last_loc = offset + int64(n)
	return buf

}

func (cf *CachedFile) getFile() *os.File {
	if (cf.fp == nil) {
		fp, err := os.Open(cf.srcfile_path)
		if (err != nil) {
			log.Fatal(err)
		}
		cf.fp = fp
	}
	return cf.fp
}

func loc_in_block(loc int64) BlockNum {
	return BlockNum(loc / BLOCK_SIZE);
}