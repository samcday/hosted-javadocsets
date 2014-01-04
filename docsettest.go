package main

import (
	"os"

	log "github.com/cihub/seelog"
	"github.com/samcday/hosted-javadocsets/docset"
)

func main() {
	defer log.Flush()
	if err := os.Remove("/tmp/woot.tgz"); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	f, err := os.Create("/tmp/woot.tgz")
	if err != nil {
		panic(err)
	}

	if err := docset.Create("com.google.inject", "guice", "3.0", f); err != nil {
		panic(err)
	}
}
