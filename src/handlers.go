// Copyright 2013 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
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
