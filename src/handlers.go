// Copyright 2010-2014 restmq authors.  All rights reserved.
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

func (s *httpServer) route() {
	// Static file server.
	http.Handle("/static/", http.FileServer(http.Dir(s.config.DocumentRoot)))
	http.Handle("/dashboard/", http.FileServer(http.Dir(s.config.DocumentRoot)))

	// Other handlers.
	http.HandleFunc("/", s.indexHandler)

	// Queue handlers
	http.HandleFunc("/q/", s.queueHandler)
	http.HandleFunc("/c/", s.cometHandler)
	http.HandleFunc("/sse/", s.sseHandler)
	http.Handle("/ws/", websocket.Handler(s.websocketHandler))

	// Management dashboard and control APIs
	http.HandleFunc("/api/policy/", s.policyHandler)
	http.HandleFunc("/api/pause/", s.pauseHandler)
	http.HandleFunc("/api/start/", s.startHandler)
	http.HandleFunc("/api/status/", s.statusHandler)
	http.HandleFunc("/api/serverstatus/", s.serverStatusHandler)
}

func (s *httpServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	qs, err := s.rmq.ListQueues()
	if err != nil {
		http.Error(w, http.StatusText(503), 503)
		context.Set(r, "log", err)
		return
	}
	for _, e := range qs {
		fmt.Fprintf(w, e)
	}
}

func (s *httpServer) queueHandler(w http.ResponseWriter, r *http.Request) {
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
		item, err := s.rmq.Get(qn, soft)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "log", err)
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
		item, err := s.rmq.Add(qn, v)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "log", err)
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
func (s *httpServer) policyHandler(w http.ResponseWriter, r *http.Request) {
	qn := r.URL.Path[len("/api/policy/"):]
	if !queueRe.MatchString(qn) {
		http.Error(w, "Invalid queue name", 400)
		return
	}
	switch r.Method {
	case "GET":
		policy, err := s.rmq.Policy(qn)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "log", err)
			return
		}
		fmt.Fprintf(w, "%s\r\n", policy)
	case "POST":
		err := s.rmq.SetPolicy(qn, r.FormValue("set"))
		if err != nil {
			if err == restmq.InvalidQueuePolicy {
				http.Error(w, err.Error(), 400)
				return
			}
			http.Error(w, http.StatusText(503), 503)
			context.Set(r, "log", err)
			return
		}
		fmt.Fprintf(w, "OK\r\n")
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, http.StatusText(405), 405)
	}
}

// /api/pause/<queuename> - pause queue streaming and hold back messages
func (s *httpServer) pauseHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *httpServer) startHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *httpServer) statusHandler(w http.ResponseWriter, r *http.Request) {
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
func (s *httpServer) serverStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, http.StatusText(405), 405)
		return
	}
}

// Streamers
func (s *httpServer) cometHandler(w http.ResponseWriter, r *http.Request) {
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
		context.Set(r, "log", "Chunked responses not supported")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	f.Flush()
	s.rmq.AddClient(qn, w)
	c, e := s.rmq.Join(qn, 30, false)
	j := json.NewEncoder(w)
L:
	for {
		select {
		case item := <-c:
			j.Encode(item)
			f.Flush()
		case err := <-e:
			context.Set(r, "log", err)
			break L
		case <-w.(http.CloseNotifier).CloseNotify():
			break L
		}
	}
}

func (s *httpServer) sseHandler(w http.ResponseWriter, r *http.Request) {
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
	s.rmq.AddClient(qn, rw) // needs to repackage it into sse proto
	c, e := s.rmq.Join(qn, 30, false)
L:
	for {
		select {
		case item := <-c:
			if sse.SendEvent(rw, &sse.MessageEvent{Data: item.JSON()}) != nil {
				break L
			}
		case err := <-e:
			context.Set(r, "log", err)
			break L
		}
	}
}

func (s *httpServer) websocketHandler(ws *websocket.Conn) {
	r := ws.Request()
	qn := r.URL.Path[len("/ws/"):]
	if !queueRe.MatchString(qn) {
		ws.Close()
	}
	s.rmq.AddClient(qn, ws)
	c, e := s.rmq.Join(qn, 600, false)
	j := json.NewEncoder(ws)
L:
	for {
		select {
		case item := <-c:
			if j.Encode(item) != nil {
				break L
			}
		case err := <-e:
			context.Set(r, "log", err)
			break L
		}
	}
}
