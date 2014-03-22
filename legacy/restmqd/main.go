// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/fiorix/go-web/httpxtra"
	"github.com/fiorix/go-web/remux"

	"bitbucket.org/gleicon/restmq/restmq"
)

var VERSION = "2.0.0"
var cfg ConfigData
var rmq *restmq.RestMQ
var queues *QueueGroup

func main() {
	cf := flag.String("config", "restmqd.xml", "set config file")
	flag.Parse()
	if err := ReadConfig(*cf, &cfg); err != nil {
		log.Fatal(err)
	}
	queues = NewQueueGroup()
	// Use all CPUs available.
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
	qn := "([a-zA-Z0-9]+)$"
	remux.HandleFunc("^/q/"+qn, QueueHandler)
	remux.HandleFunc("^/c/"+qn, CometQueueHandler)
	//remux.HandleFunc("^/ws/"+qn, WebSocketQueueHandler)
	http.Handle("/ws/", websocket.Handler(WebSocketQueueHandler))
	server := http.Server{
		Handler: httpxtra.Handler{
			Handler: remux.DefaultServeMux,
			Logger:  logger,
		},
	}
	wg := &sync.WaitGroup{}
	if cfg.Addr != "" {
		wg.Add(1)
		http := server
		http.Addr = cfg.Addr
		log.Printf("Starting HTTP server on %s", http.Addr)
		go func() {
			log.Fatal(http.ListenAndServe())
			wg.Done()
		}()
	}
	if cfg.SSL.Addr != "" {
		wg.Add(1)
		https := server
		https.Addr = cfg.SSL.Addr
		log.Printf("Starting HTTPS server on %s", https.Addr)
		go func() {
			log.Fatal(https.ListenAndServeTLS(
				cfg.SSL.CertFile, cfg.SSL.KeyFile))
			wg.Done()
		}()

	}
	wg.Wait()
}

func logger(r *http.Request, created time.Time, status, bytes int) {
	fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))
}
