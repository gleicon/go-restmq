// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/fiorix/go-web/http"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "tinymq")
}

func QueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Vars[0] == "" { // queue name
		http.Error(w, "Queue Not Found. Use /q/queue_name", 404)
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
		item, err := rmq.Get(r.Vars[0], soft)
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
		item, err := rmq.Add(r.Vars[0], v)
		if err != nil {
			http.Error(w, http.StatusText(503), 503)
			return
		}
		s := item.String()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, s)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, http.StatusText(405), 405)
	}
}

/*
func QueueHandler(w http.ResponseWriter, r *http.Request) {
	queue_name := r.Vars[0]
	if queue_name == "" {
		http.Error(w, "Queue name not given", 404)
		return
	}
	qname := queue_name + QUEUE_SUFFIX
	quuid := queue_name + UUID_SUFFIX
	soft := r.FormValue("soft")
	callback := r.FormValue("callback")

	switch r.Method {
	case "GET":
		var (
			p   string
			err error
		)

		if soft != "" {
			p, err = redis_client.LIndex(qname, -1)
		} else {
			p, err = redis_client.RPop(qname)
		}

		if p == "" {
			http.Error(w, "Empty queue", 404)
			return
		}

		if err == nil {
			v, err := redis_client.Get(p)
			if err == nil {
				resp := Response{"key": p, "value": v}

				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					if callback == "" {
						fmt.Fprintf(w, "%s\n", resp)
					} else {
						fmt.Fprintf(w, "%s(%s);\n", callback, resp)
					}
					return
				}
				http.Error(w, "Value not found", 500)
				return
			}
		}
	case "POST":
		value := r.FormValue("value")
		if value == "" {
			http.Error(w, "Empty value", 401)
		}

		uuid, _ := redis_client.Incr(quuid)
		lkey := queue_name + ":" + strconv.Itoa(uuid)
		redis_client.Set(lkey, value)
		redis_client.LPush(qname, lkey)
		resp := Response{"key": lkey, "value": value}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%s\n", resp)
	default:
		http.Error(w, "Method not accepted", 400)
		return

	}
}
*/
