// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/fiorix/go-web/http"
	"github.com/fiorix/go-web/sse"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "RestMQ v%s\n", VERSION)
}

func QueueHandler(w http.ResponseWriter, r *http.Request) {
	queue := r.Vars[0]
	switch r.Method {
	case "GET":
		var soft bool
		if r.FormValue("soft") == "" {
			soft = false
		} else {
			soft = true
		}
		item, err := rmq.Get(queue, soft)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			return
		}
		s := item.String()
		if s == "" {
			http.Error(w, "Empty queue", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, s)
	case "POST":
		v := r.FormValue("value")
		if v == "" {
			http.Error(w, "Empty value", 400)
			return
		}
		item, err := rmq.Add(queue, v)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			return
		}
		s := item.String()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, s)
		queues.Say(queue, s)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, http.StatusText(405), 405)
	}
}

func CometQueueHandler(w http.ResponseWriter, r *http.Request) {
	conn, buf, err := sse.ServeEvents(w)
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer conn.Close()
	cs := queues.Join(r.Vars[0])
	for s := range cs {
		err = sse.SendEvent(buf, &sse.MessageEvent{Data: s})
		if err != nil {
			break
		}
	}
	queues.Part(r.Vars[0], cs)
}

func WebSocketQueueHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: add/get
}
