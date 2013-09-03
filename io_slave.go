package smoothfs

import (
	"log"
	"time"
)

const IO_RESPONSE_TIMEOUT = 60 * time.Second

func io_slave(fs *SmoothFS, id int, ch chan WorkEntry) {
	log.Printf("Starting io slave %d", id)
	for req := range ch {
		req.Process()
		responder := req.Responder()
		if responder != nil {
			select {
			case responder <- req:
				log.Printf("Slave %d Sent read response for %s", id, req)
			case <-time.After(IO_RESPONSE_TIMEOUT):
				log.Printf("Slave %d IO response for %s timed out for some reason", id, req)
			}
		}
	}
	log.Printf("Closing io slave %d", id)
}
