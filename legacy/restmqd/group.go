// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"sync"
)

type Client chan string

// Group is the interface of a group of clients.
type Group interface {
	Join() Client
	Part(client Client)
	Say(msg string) // Broadcast the msg to all clients
	// TODO: SayOne for queues with round-robin policy
}

// ClientGroup is our thread-safe Group.
type ClientGroup struct {
	lk      sync.RWMutex
	clients map[Client]bool
}

// NewClientGroups creates and initializes a new Group.
func NewClientGroup() *ClientGroup {
	c := ClientGroup{}
	c.clients = make(map[Client]bool)
	return &c
}

// Join adds a client to the Group.
func (cg *ClientGroup) Join() Client {
	c := make(Client)
	cg.lk.Lock()
	defer cg.lk.Unlock()
	cg.clients[c] = true
	return c
}

// Part removes a client from the Group.
func (cg *ClientGroup) Part(c Client) {
	cg.lk.RLock()
	_, exists := cg.clients[c]
	cg.lk.RUnlock()
	if exists {
		cg.lk.Lock()
		defer cg.lk.Unlock()
		delete(cg.clients, c)
		close(c)
	}
}

// Say broadcasts a message to all clients in the Group.
func (cg *ClientGroup) Say(msg string) {
	cg.lk.RLock()
	defer cg.lk.RUnlock()
	for c, _ := range cg.clients {
		c <- msg
	}
}

// QueueGroup implements one Group per queue.
// QueueGroup manages HTTP clients subscribed to a queue for live updates.
type QueueGroup struct {
	lk     sync.RWMutex
	queues map[string]Group
}

// NewQueueGroup creates and initializes a Group of queues.
func NewQueueGroup() *QueueGroup {
	cg := QueueGroup{}
	cg.queues = make(map[string]Group)
	return &cg
}

// Join adds the HTTP client to the Group of a particular queue.
func (cg *QueueGroup) Join(queue string) Client {
	var q Group
	var exists bool
	cg.lk.RLock()
	q, exists = cg.queues[queue]
	cg.lk.RUnlock()
	if !exists {
		cg.lk.Lock()
		defer cg.lk.Unlock()
		q = NewClientGroup()
		cg.queues[queue] = q
	}
	return q.Join()
}

// Part removes the HTTP client from the Group of a particular queue.
func (cg *QueueGroup) Part(queue string, c Client) {
	cg.lk.RLock()
	if q, exists := cg.queues[queue]; exists {
		q.Part(c)
	}
}

// Say broadcasts a message to all clients in the Group of a particular queue.
func (cg *QueueGroup) Say(queue, msg string) {
	cg.lk.RLock()
	q, exists := cg.queues[queue]
	cg.lk.RUnlock()
	if exists {
		q.Say(msg)
	}
}
