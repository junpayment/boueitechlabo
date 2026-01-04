package main

import (
	"boueitechlabo/bouteitech"
	"log"
	"net"
	"os"
)

const (
	envCACert = "CERT_PATH"
	envCAKey  = "KEY_PATH"
)

func main() {
	signer, err := bouteitech.NewSigner(os.Getenv(envCACert), os.Getenv(envCAKey))
	if err != nil {
		log.Fatal(err)
	}
	ln, _ := net.Listen("tcp", ":8080")
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go bouteitech.HandleConn(c, signer)
	}
}
