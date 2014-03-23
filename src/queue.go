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

// for each queue:
//  an entry into QUEUES set
//  a hash with UUID and REFCOUNT members
//  a list
// for each message in queue:
//  incr hash:UUID
//  an element with the payload
//  an entry into the list with the element key

package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fiorix/go-redis/redis"
)

const (
	QUEUEINDEX          = "restmq:queue:index"
	QUEUEMETAHASHPREFIX = "restmq:queue:meta:%s"
	QUEUEUUID           = "uuid"
	QUEUEREFCOUNTER     = "refcounter"
)

// Queue is the RestMQ queue handler.
type Queue struct {
	rc *redis.Client
	qi []string // queue index
}

type Item struct {
	Id    int    `json:"id"`
	Count int    `json:"count"`
	Value string `json:"value"`
}

func (i *Item) Write(w http.ResponseWriter) error {
	return writeJSON(w, i)
}

// newQueue creates, initializes and returns a new instance of RestMQ.
// RestMQ instances are safe for concurrent access.
func newQueue(addr ...string) *Queue {
	q := Queue{redis.New(addr...), []string{}}
	go func() {
		for {
			log.Print("Refreshing index cache")
			qi, err := q.rc.SMembers(QUEUEINDEX)
			if err != nil {
				log.Print("Error fetching queue index" + err.Error())
			}
			q.qi = qi
			time.Sleep(60000 * time.Millisecond)
		}
	}()
	return &q
}

// if queue is not in the index, add it.
// any race condition between refreshing the set + adding it can be tuned
// by the goroutine sleep
// I promise it wont be too expensive.
func (mq *Queue) checkAndAddToIndex(queue string) {
	found := false
	for _, i := range mq.qi {
		if i == queue {
			found = true
		}
	}

	if !found {
		mq.rc.SAdd(QUEUEINDEX, queue)
	}
}

// Add adds one item into the given queue, which is created on demand.
func (mq *Queue) Add(queue, value string) (*Item, error) {
	// TODO: Fix for cases when Redis disconnects in between the commands.

	// TODO: think better about when/how to include a queue in the index
	mq.checkAndAddToIndex(queue)
	n, err := mq.rc.HIncrBy(queueMetadata(queue), QUEUEUUID, 1)
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
	return &Item{n, 0, value}, nil
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
		err      error
		ns       string
		qn       = queueName(queue)
		refcount = 0
	)
	if soft {
		ns, err = mq.rc.LIndex(qn, -1)
		if err != nil {
			return nil, err // This causes HTTP 503
		}
		refcounts, err := mq.rc.HGet(queueMetadata(queue), QUEUEREFCOUNTER)
		if err != nil {
			return nil, err // This causes HTTP 503
		}
		refcount, err = strconv.Atoi(refcounts)
		if err != nil {
			return nil, err // This causes HTTP 503
		}
	} else {
		ns, err = mq.rc.RPop(qn)
		if err != nil {
			return nil, err // This causes HTTP 503
		}
		err := mq.rc.HSet(queueMetadata(queue), QUEUEREFCOUNTER, "0")
		if err != nil {
			return nil, err // This causes HTTP 503
		}
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
	return &Item{n, refcount, v}, nil
}

func (mq *Queue) Del(queue string) error {
	_, err := mq.Get(queue, false)
	return err
}

// Len returns the length of the given queue.
func (mq *Queue) Len(queue string) (int, error) {
	l, err := mq.rc.LLen(queueName(queue))
	if err != nil {
		return -1, err // This causes HTTP 503
	}
	return l, nil
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
// TO BE REMOVED
func (mq *Queue) Tail(queue string) error {
	return nil
}

// CountElements does something different than Len.
// TODO: Gleicon
// to be obsoleted, we need transactions/track all garbage for each queue item
func (mq *Queue) CountElements(queue string) (int, error) {
	return 0, nil
}

// LastItems returns the last n items in the queue.
// TO BE OBSOLETED, QUEUE IS NOT A DATABASE
func (mq *Queue) LastItems(queue string, n int) ([]string, error) {
	return nil, nil
}

func queueName(name string) string {
	return "q:" + name
}

func queueMetadata(name string) string {
	return fmt.Sprintf(QUEUEMETAHASHPREFIX, name)
}

func queueCount(name string) string {
	return "n:" + name
}
