package smoothfs


func io_slave(c chan IOReq) {
	for {
		select {
		case req, ok := (<-c):
			if (ok) {
				// have data
				req.Node.ReadAll(nil) // XXX
			} else {
				return
			}
		}
	}
}