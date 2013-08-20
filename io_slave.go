package smoothfs

import (
	"log"
)


func io_slave(fs *SmoothFS, id int, c chan IOReq) {
	log.Printf("Starting io slave %d", id)
	for {
		select {
		case req, ok := (<-c):
			if (ok) {
				// have data
				req.Node.Attr() // XXX
			} else {
				log.Printf("Closing io slave %d", id)
				return
			}
		}
	}
}
