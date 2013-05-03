// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/remux"

	"bitbucket.org/gleicon/restmq/restmq"
)

var VERSION = "2.0.0"
var rmq *restmq.RestMQ

func main() {
	numCPU := runtime.NumCPU()
	label := "CPU"
	if numCPU > 1 {
		label += "s"
	}
	runtime.GOMAXPROCS(numCPU)
	log.Printf("restmqd v%s (%d %s)", VERSION, numCPU, label)
	// Create the global RestMQ instance.
	rmq = restmq.New("127.0.0.1:6379")
	// HTTP handlers
	remux.HandleFunc("^/$", IndexHandler)
	remux.HandleFunc("^/q/([a-zA-Z0-9]+)$", QueueHandler)
	server := http.Server{
		Handler: remux.DefaultServeMux,
		Logger:  logger,
	}
	wg := &sync.WaitGroup{}
	// TODO: Parse the config file and check whether to startup HTTP
	if true {
		wg.Add(1)
		http := server
		http.Addr = ":8080" // TODO: Use from config file
		log.Printf("Starting HTTP server on %s", http.Addr)
		go func() {
			log.Fatal(http.ListenAndServe())
			wg.Done()
		}()
	}
	// TODO: Parse the config file and check whether to startup HTTPS
	if false {
		wg.Add(1)
		https := server
		https.Addr = ":8443"
		log.Printf("Starting HTTP server on %s", https.Addr)
		go func() {
			log.Fatal(https.ListenAndServeTLS("", ""))
			wg.Done()
		}()

	}
	wg.Wait()
}

func logger(w http.ResponseWriter, r *http.Request) {
	var s string
	if r.TLS != nil {
		s = "S" // soz no ternary :/
	}
	log.Printf("HTTP%s %d %s %s (%s) :: %s",
		s,
		w.Status(),
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		time.Since(r.Created))
}
