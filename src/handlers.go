// Copyright 2013 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/context"
)

var queueRe = regexp.MustCompile("^([a-zA-Z0-9]+)$")

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, world\r\n")
}

func QueueHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/q/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	switch r.Method {
	case "GET":
		var soft bool
		if r.FormValue("soft") == "" {
			soft = false
		} else {
			soft = true
		}
		item, err := RestMQ.Get(qn, soft)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "info", err)
			return
		} else if item == nil {
			http.Error(w, "Queue is empty", 404)
			return
		}
		item.Write(w)
	case "POST":
		v := r.FormValue("value")
		if v == "" {
			http.Error(w, "Missing 'value=' argument", 400)
			return
		}
		item, err := RestMQ.Add(qn, v)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "info", err)
			return
		}
		item.Write(w)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, http.StatusText(405), 405)
	}
}

func CometHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/c/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(405), 405)
		return
	}
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, http.StatusText(503), 503)
		context.Set(r, "info", "Chunked responses not supported")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	f.Flush()
	c, e := RestMQ.Join(qn, 30)
	j := json.NewEncoder(w)
L:
	for {
		select {
		case item := <-c:
			j.Encode(item)
			f.Flush()
		case err := <-e:
			context.Set(r, "info", err)
			break L
		case <-w.(http.CloseNotifier).CloseNotify():
			break L
		}
	}
}