// Copyright 2013 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	html "html/template"
	text "text/template"

	"github.com/gleicon/go-restmq/restmq"
	"github.com/gleicon/go-restmq/restmq/redis"
)

var (
	VERSION = "tip"

	Config *ConfigData

	// Templates
	HTML *html.Template
	TEXT *text.Template

	// DBs
	RestMQ restmq.Queue
)

func main() {
	Configfile := flag.String("c", "restmq.conf", "set config file")
	flag.Usage = func() {
		fmt.Println("Usage: restmq [-c restmq.conf]")
		os.Exit(1)
	}
	flag.Parse()

	var err error
	Config, err = LoadConfig(*Configfile)
	if err != nil {
		log.Fatal(err)
	}

	// Parse templates.
	HTML = html.Must(html.ParseGlob(Config.TemplatesDir + "/*.html"))
	TEXT = text.Must(text.ParseGlob(Config.TemplatesDir + "/*.txt"))

	// Set up databases.
	RestMQ = restmq_redis.New(Config.DB.Redis)

	// Set GOMAXPROCS and show server info.
	var cpuinfo string
	if n := runtime.NumCPU(); n > 1 {
		runtime.GOMAXPROCS(n)
		cpuinfo = fmt.Sprintf("%d CPUs", n)
	} else {
		cpuinfo = "1 CPU"
	}
	log.Printf("restmqd v%s (%s)", VERSION, cpuinfo)

	// Add routes, and run HTTP and HTTPS servers.
	routeHTTP()
	if Config.HTTP.Addr != "" {
		go listenHTTP()
	}
	if Config.HTTPS.Addr != "" {
		go listenHTTPS()
	}
	select {}
}
