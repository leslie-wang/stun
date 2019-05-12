package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gortc/stun"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, os.Args[0], "stun.l.google.com:19302")
	}
	flag.Parse()
	addr := flag.Arg(0)
	if addr == "" {
		addr = "stun.l.google.com:19302"
	}
	c, err := stun.Dial("udp", addr)
	if err != nil {
		log.Fatal("dial:", err)
	}

	var xorAddr stun.XORMappedAddress
	ch := make(chan struct{})
	if err := c.Start(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
		if res.Error != nil {
			log.Fatalln(err)
		}
		for _, a := range res.Message.Attributes {
			fmt.Println(a)
		}
		if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
			log.Fatalln(getErr)
		}
		fmt.Printf("http://[%s]:%d\n", xorAddr.IP.String(), xorAddr.Port)
		ch <- struct{}{}
	}); err != nil {
		log.Fatal("do:", err)
	}

	<-ch

	http.ListenAndServe(fmt.Sprintf(":%d", xorAddr.Port), http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(200)
		fmt.Fprintf(rw, "Hello there!\n")
		fmt.Fprintf(rw, "test-page: http://[%s]:%d\n", xorAddr.IP.String(), xorAddr.Port)
	}))

	if err := c.Close(); err != nil {
		log.Fatalln(err)
	}
}
