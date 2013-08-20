package smoothfs

import (
	"log"
)


func io_slave(fs *SmoothFS, id int, c chan IOReq) {
	log.Printf("Starting io slave %d", id)
	for {
		select {
		case req, ok := (<-c):
			if ok {
				req.CachedFile.internalRead(req)
				responder := req.Responder
				if responder != nil {
					log.Printf("Slave %d Sending read response block %d", id, req.BlockNum)
					responder <- req
				}
			} else {
				log.Printf("Closing io slave %d", id)
				return
			}
		}
	}
}
