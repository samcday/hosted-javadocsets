package main

import log "github.com/cihub/seelog"

func main() {
	defer log.Flush()
	log.Infof("Hello %d!", 123)
}
