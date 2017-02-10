package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"runtime"

	"os"

	"github.com/namsral/flag"
	"github.com/noonat/whiskey/prefork"
	"github.com/noonat/whiskey/wsgi"
)

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	var (
		addr       string
		workers    int
		wsgiConns  int
		wsgiModule string
	)
	flag.StringVar(&addr, "addr", ":8080", "Listen for HTTP connections on this address")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "Number of worker processes.")
	flag.StringVar(&wsgiModule, "wsgi-module", "", "Module and function to run for the WSGI application. (e.g. my_wsgi_app:application)")
	flag.IntVar(&wsgiConns, "wsgi-conns", 1000, "Number of simultaneous connections per worker.")
	flag.Parse()
	if wsgiModule == "" {
		fmt.Fprintln(os.Stderr, "error: -wsgi-module is required")
		flag.Usage()
		os.Exit(1)
	}

	go http.ListenAndServe(":8181", http.DefaultServeMux)

	w := &wsgi.Worker{Module: wsgiModule, NumConns: wsgiConns}
	if err := prefork.Run(w, addr, workers, logger); err != nil {
		log.Fatalf("%+v\n", err)
	}
}
