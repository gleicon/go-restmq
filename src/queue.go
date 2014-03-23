// Copyright 2009-2013 restmq authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// RestMQ protocol in Go, based on
// https://github.com/gleicon/restmq/blob/master/src/restmq/core.py
//
// TODO: update the URL above when this branch turns master.
//
// RestMQ is controlled by a handler, obtained by calling the New function.
//
// The handler currently uses Redis as the backend, but there are plans to
// support leveldb and perhaps even other storage engines. (MariaDB?)

package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/fiorix/go-redis/redis"
)

// Queue is the RestMQ queue handler.
type Queue struct {
	rc *redis.Client
}

type Item struct {
	Id    int    `json:"id"`
	Value string `json:"value"`
}

func (i *Item) Write(w http.ResponseWriter) error {
	return writeJSON(w, i)
}

// newQueue creates, initializes and returns a new instance of RestMQ.
// RestMQ instances are safe for concurrent access.
func newQueue(addr ...string) *Queue {
	return &Queue{redis.New(addr...)}
}

// Add adds one item into the given queue, which is created on demand.
func (mq *Queue) Add(queue, value string) (*Item, error) {
	// TODO: Fix for cases when Redis disconnects in between the commands.
	n, err := mq.rc.Incr(queueCount(queue))
	if err != nil {
		return nil, err
	}
	ns := strconv.Itoa(n)
	if err = mq.rc.Set(queue+":"+ns, value); err != nil {
		return nil, err
	}
	if _, err = mq.rc.LPush(queueName(queue), ns); err != nil {
		return nil, err
	}
	return &Item{n, value}, nil
}

// Get returns one item from the given queue.
// Get can be "soft" or "hard" (hard is the default).
//
// A "soft" get is when an item is returned but remains in the queue, and
// its reference counter is incremented.
//
// The "hard" get returns an item from the queue without incrementing its
// reference counter. The item is permanently erased from the queue.
func (mq *Queue) Get(queue string, soft bool) (*Item, error) {
	var (
		err error
		ns  string
	)
	if soft {
		ns, err = mq.rc.LIndex(queueName(queue), -1)
	} else {
		ns, err = mq.rc.RPop(queueName(queue))
	}
	if err != nil {
		return nil, err // Redis error
	} else if ns == "" {
		return nil, nil // Empty queue
	}
	var v string
	v, err = mq.rc.Get(queue + ":" + ns)
	if err != nil {
		return nil, err // Redis error
	}
	n, err := strconv.Atoi(ns)
	if err != nil {
		return nil, err // This causes HTTP 503
	}
	return &Item{n, v}, nil
}

// Join returns a channel for a given queue, which pops Item items as
// they are added to the queue.
// Join behaves like an infinite Get, pushing items through the channel.
func (mq *Queue) Join(queue string) (<-chan *Item, <-chan error) {
	c := make(chan *Item)
	e := make(chan error)
	go func() {
		for {
			log.Println("calling brpop on", queue)
			_, ns, err := mq.rc.BRPop(5, queueName(queue))
			if err != nil {
				e <- err
				close(e)
				close(c)
				return
			}
			var v string
			log.Println("calling get on", queue)
			v, err = mq.rc.Get(queue + ":" + ns)
			if err != nil {
				e <- err
				close(e)
				close(c)
				return
			}
			n, err := strconv.Atoi(ns)
			if err != nil {
				e <- err
				close(e)
				close(c)
				return
			}
			c <- &Item{n, v}
		}
	}()
	return c, e
}

// GetDel is the "hard" Get. wtf? getdel == pop
func (mq *Queue) GetDel(queue string) (string, error) {
	return "", nil
}

// Del does something.
// TODO: Gleicon ^^
func (mq *Queue) Del(queue string) error {
	return nil
}

// Len returns the length of the given queue.
func (mq *Queue) Len(queue string) (int, error) {
	return 0, nil
}

// All returns all items in the given queue.
func (mq *Queue) All(queue string) ([]string, error) {
	return nil, nil
}

// Policy returns the current policy of the given queue.
func (mq *Queue) Policy(queue string) (string, error) {
	return "", nil
}

// PolicySet sets a new policy for the given queue.
func (mq *Queue) PolicySet(queue string) error {
	return nil
}

// Tail follows a queue.
// TODO: Gleicon update
// TODO: Channel?
func (mq *Queue) Tail(queue string) error {
	return nil
}

// CountElements does something different than Len.
// TODO: Gleicon
func (mq *Queue) CountElements(queue string) (int, error) {
	return 0, nil
}

// LastItems returns the last n items in the queue.
func (mq *Queue) LastItems(queue string, n int) ([]string, error) {
	return nil, nil
}

func queueName(name string) string {
	return "q:" + name
}

func queueCount(name string) string {
	return "n:" + name
}
