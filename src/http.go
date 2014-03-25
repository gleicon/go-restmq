// Copyright 2010-2014 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"code.google.com/p/go.net/websocket"
	"github.com/fiorix/go-web/httpxtra"
	"github.com/gorilla/context"
)

func routeHTTP() {
	// Static file server.
	http.Handle("/static/", http.FileServer(http.Dir(Config.DocumentRoot)))
	http.Handle("/dashboard/", http.StripPrefix("/dashboard/", http.FileServer(http.Dir(Config.DocumentRoot+"/dashboard/"))))
	log.Println(http.Dir(Config.DocumentRoot + "/dashboard/"))

	// Public handlers: add your own
	http.HandleFunc("/", IndexHandler)

	// Queue handlers
	http.HandleFunc("/q/", QueueHandler)
	http.HandleFunc("/c/", CometHandler)
	http.HandleFunc("/sse/", SSEHandler)
	http.Handle("/ws/", websocket.Handler(WebsocketHandler))

	// Management dashboard and control APIs
	http.HandleFunc("/api/policy/", PolicyHandler)
	http.HandleFunc("/api/pause/", PauseHandler)
	http.HandleFunc("/api/start/", StartHandler)
	http.HandleFunc("/api/status/", StatusHandler)
	http.HandleFunc("/api/serverstatus/", ServerStatusHandler)
}

func listenHTTP() {
	s := http.Server{
		Addr:    Config.HTTP.Addr,
		Handler: httpxtra.Handler{Logger: httpLogger},
	}
	log.Println("Starting HTTP server on", Config.HTTP.Addr)
	log.Fatal(s.ListenAndServe())
}

func listenHTTPS() {
	s := http.Server{
		Addr:    Config.HTTPS.Addr,
		Handler: httpxtra.Handler{Logger: httpLogger},
	}
	log.Println("Starting HTTPS server on", Config.HTTPS.Addr)
	log.Fatal(s.ListenAndServeTLS(Config.HTTPS.CertFile, Config.HTTPS.KeyFile))
}

// httpError renders the default error message based on
// the status code, and sets the "log" context variable with the error.
func httpError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	// TODO: render error page instead of text?
	http.Error(w, http.StatusText(code), code)

	if msg != "" {
		context.Set(r, "log", msg)
	}
}

// httpLogger is called at the end of every HTTP request. It dumps one
// log line per request.
//
// The "log" context variable can be used to add extra information to
// the logging, such as database or template errors.
func httpLogger(r *http.Request, created time.Time, status, bytes int) {
	//fmt.Println(httpxtra.ApacheCommonLog(r, created, status, bytes))

	var proto, msg string

	if r.TLS == nil {
		proto = "HTTP"
	} else {
		proto = "HTTPS"
	}

	if tmp := context.Get(r, "log"); tmp != nil {
		msg = fmt.Sprintf(" (%s)", tmp)
		context.Clear(r)
	}

	log.Printf("%s %d %s %q (%s) :: %d bytes in %s%s",
		proto,
		status,
		r.Method,
		r.URL.Path,
		remoteIP(r),
		bytes,
		time.Since(created),
		msg,
	)
}
