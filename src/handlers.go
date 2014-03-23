// Copyright 2013 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/fiorix/go-web/sse"
	"github.com/gorilla/context"

	"github.com/gleicon/go-restmq/restmq"
)

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
		item.WriteJSON(w)
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
		item.WriteJSON(w)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, http.StatusText(405), 405)
	}
}

// Management and queue control APIs

// /api/policy/<queuename> - switch between broadcast and roundrobin delivery
// for streaming methods
func PolicyHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/api/policy/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	switch r.Method {
	case "GET":
		policy, err := RestMQ.Policy(qn)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "info", err)
			return
		}
		fmt.Fprintf(w, "%s\r\n", policy)
	case "POST":
		err := RestMQ.SetPolicy(qn, r.FormValue("set"))
		if err != nil {
			if err == restmq.InvalidQueuePolicy {
				http.Error(w, err.Error(), 400)
				return
			}
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "info", err)
			return
		}
		fmt.Fprintf(w, "OK\r\n")
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, http.StatusText(405), 405)
	}
}

// /api/pause/<queuename> - pause queue streaming and hold back messages
func PauseHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/api/pause/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	if r.Method != "POST" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(405), 405)
		return
	}
}

// /api/start/<queuename> - start queue streaming
func StartHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/api/start/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	if r.Method != "POST" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(405), 405)
		return
	}
}

// /api/status/<queuename> - status for a given queue
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/api/status/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(405), 405)
		return
	}
}

// /api/serverstatus - server status
func ServerStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(405), 405)
		return
	}
}

// Streamers
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

func SSEHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/sse/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(405), 405)
		return
	}
	conn, rw, err := sse.ServeEvents(w)
	if err != nil {
		conn.Close()
		return
	}
	defer conn.Close()
	c, e := RestMQ.Join(qn, 30)
L:
	for {
		select {
		case item := <-c:
			if sse.SendEvent(rw, &sse.MessageEvent{Data: item.JSON()}) != nil {
				break L
			}
		case err := <-e:
			context.Set(r, "info", err)
			break L
		}
	}
}

func WebsocketHandler(ws *websocket.Conn) {
	r := ws.Request()
	qn := r.URL.Path[len("/ws/"):]
	if !queueRe.MatchString(qn) {
		ws.Close()
	}
	c, e := RestMQ.Join(qn, 600)
	j := json.NewEncoder(ws)
L:
	for {
		select {
		case item := <-c:
			if j.Encode(item) != nil {
				break L
			}
		case err := <-e:
			context.Set(r, "info", err)
			break L
		}
	}
}
