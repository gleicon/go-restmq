// Copyright 2010-2014 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

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

package rmq

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/fiorix/go-redis/redis"
	"github.com/gleicon/go-restmq/restmq"
)

const (
	QueueIndex          = "restmq:queue:index"
	QueueMetaHashPrefix = "restmq:queue:meta:%s"
	QueueNamePrefix     = "restmq:queue:name:%s"
	QueueUUID           = "uuid"
	QueueRefCount       = "refcount"
	QueuePolicy         = "policy"
	DefaultPolicy       = "fifo"
)

// Queue is the RestMQ queue handler.
type Queue struct {
	rc *redis.Client
	cp *restmq.ClientPresence

	qimu sync.Mutex // protects qi
	qi   []string   // queue index
}

// New creates, initializes and returns a new instance of RestMQ.
// RestMQ instances are safe for concurrent access.
func New(addr string) *Queue {
	mq := &Queue{
		rc: redis.New(addr),
		cp: restmq.NewClientPresence(),
	}
	go mq.runIndex(30 * time.Second)
	return mq
}

func (mq *Queue) runIndex(interval time.Duration) {
	var (
		err error
		qi  []string
	)
	for {
		if qi, err = mq.rc.SMembers(QueueIndex); err != nil {
			log.Print("Error fetching queue index", err)
		} else {
			log.Print("Queue index cache updated")
			mq.qimu.Lock()
			mq.qi = qi
			mq.qimu.Unlock()
		}
		time.Sleep(interval)
	}
}

// if queue is not in the index, add it.
// any race condition between refreshing the set + adding it can be tuned
// by the goroutine sleep
// I promise it wont be too expensive.
func (mq *Queue) updateIndex(queue string) {
	mq.qimu.Lock()
	defer mq.qimu.Unlock()
	for _, name := range mq.qi {
		if name == queue {
			return // queue already indexed, bye
		}
	}
	mq.rc.SAdd(QueueIndex, queue)
}

// Add adds one item into the given queue, which is created on demand.
func (mq *Queue) Add(queue, value string) (*restmq.Item, error) {
	// TODO: Fix for cases when Redis disconnects in between the commands.

	// TODO: think better about when/how to include a queue in the index
	mq.updateIndex(queue)
	n, err := mq.rc.HIncrBy(queueMetadata(queue), QueueUUID, 1)
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
	return &restmq.Item{n, 0, value}, nil
}

// Get returns one item from the given queue.
// Get can be "soft" or "hard" (hard is the default).
//
// A "soft" get is when an item is returned but remains in the queue, and
// its reference counter is incremented.
//
// The "hard" get returns an item from the queue without incrementing its
// reference counter, and the item is permanently erased from the queue.
func (mq *Queue) Get(queue string, soft bool) (*restmq.Item, error) {
	var (
		err error
		ns  string
		r   int
		qn  = queueName(queue)
	)
	if soft {
		ns, err = mq.rc.LIndex(qn, -1)
		if err != nil {
			return nil, err // This causes HTTP 503
		}
		r, err = mq.rc.HIncrBy(queueMetadata(queue), QueueRefCount, 1)
		if err != nil {
			return nil, err // This causes HTTP 503
		}
	} else {
		ns, err = mq.rc.RPop(qn)
		if err != nil {
			return nil, err // This causes HTTP 503
		}
		err = mq.rc.HSet(queueMetadata(queue), QueueRefCount, "0")
		if err != nil {
			return nil, err // This causes HTTP 503
		}
	}
	if ns == "" {
		return nil, nil // Empty queue
	}
	var v string
	v, err = mq.rc.Get(queue + ":" + ns)
	if err != nil {
		return nil, err // Redis error
	}
	if !soft {
		mq.rc.Del(queue + ":" + ns)
	}
	var n int
	if n, err = strconv.Atoi(ns); err != nil {
		return nil, err // This causes HTTP 503
	}
	return &restmq.Item{n, r, v}, nil
}

// Join hooks into the queue in a goroutine, and returns channels which
// get items as they're added to the queue, or errors.
//
// If multiple clients call Join for the same queue, the first client to be
// served is the one that was waiting for more time (the first that
// called Join for the queue).
func (mq *Queue) Join(queue string, timeout int, soft bool) (<-chan *restmq.Item, <-chan error) {
	c := make(chan *restmq.Item)
	e := make(chan error)
	go func() {
		var (
			rc int
		)
		for {
			_, ns, err := mq.rc.BRPop(timeout, queueName(queue))
			if err != nil {
				e <- err
				return
			}
			n, err := strconv.Atoi(ns)
			if err != nil {
				e <- err
				return
			}
			var v string
			v, err = mq.rc.Get(queue + ":" + ns)
			if err != nil {
				e <- err
				return
			}
			if soft {
				rc, err = mq.rc.HIncrBy(queueMetadata(queue), QueueRefCount, 1)
				if err != nil {
					e <- err
					return
				}
			} else {
				rs, err := mq.rc.HGet(queueMetadata(queue), QueueRefCount)
				if err != nil {
					e <- err
					return
				}
				if rs != "" {
					rc, err = strconv.Atoi(rs)
					if err != nil {
						e <- err
						return
					}
				}
				err = mq.rc.HSet(queueMetadata(queue), QueueRefCount, "0")
				if err != nil {
					e <- err
					return
				}
			}
			c <- &restmq.Item{n, rc, v}
		}
	}()
	return c, e
}

// populates presence with each client io.Writer
func (mq *Queue) AddClient(queue string, w io.Writer) {
	mq.cp.Add(queue, w)
}

// GetDel is the "hard" Get.
func (mq *Queue) GetDel(queue string) (string, error) {
	return "", nil
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
	policy, err := mq.rc.HGet(queueMetadata(queue), QueuePolicy)
	if err != nil {
		return "", err
	}
	if policy == "" {
		return DefaultPolicy, nil
	}
	return policy, nil
}

// PolicySet sets a new policy for the given queue.
func (mq *Queue) SetPolicy(queue string, policy string) error {
	switch policy {
	case "fifo":
	case "broadcast":
	default:
		return restmq.InvalidQueuePolicy
	}
	return mq.rc.HSet(queueMetadata(queue), QueuePolicy, policy)
}

func (mq *Queue) ListQueues() ([]string, error) {
	return mq.rc.SMembers(QueueIndex)
}

func (mq *Queue) Pause() error {
	return nil
}

func (mq *Queue) Start() error {
	return nil
}

func queueName(name string) string {
	return fmt.Sprintf(QueueNamePrefix, name)
}

func queueMetadata(name string) string {
	return fmt.Sprintf(QueueMetaHashPrefix, name)
}
