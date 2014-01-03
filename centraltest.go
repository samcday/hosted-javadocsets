package main

import (
	_ "fmt"
	"io"
	"os"

	"github.com/samcday/hosted-javadocsets/mavencentral"
)

func main() {
	r, err := mavencentral.GetArtifact("com.google.inject", "guice", "3.0", "javadoc")
	if err != nil {
		panic(err)
	}
	defer r.Close()
	file, err := os.Create("/tmp/dl")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	io.Copy(file, r)
}
