// Copyright 2010-2014 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// RestMQ protocol in Go, based on
// https://github.com/gleicon/restmq/blob/master/src/restmq/core.py

package restmq

import (
	"bufio"
	"errors"
	"net/http"

	"code.google.com/p/go.net/websocket"
)

type Queue interface {
	// Add new item into the queue.
	Add(queue, value string) (*Item, error)

	// Get gets an item from the queue. Optional Soft argument makes it
	// keep the item in the queue. The default is to delete it.
	Get(queue string, soft bool) (*Item, error)

	// Join hooks up into a queue and sends new items to the channel
	// as they're added to the queue. Multiple callers receive items
	// distributed as per the queue's policy.
	//
	// See the queue policy documentation for details.
	Join(queue string, timeout int, soft bool) (<-chan *Item, <-chan error)

	// Policy returns the current queue policy. If the queue does not
	// exist, it is created and the default policy is used.
	Policy(queue string) (string, error)

	// SetPolicy sets the queue policy. If the queue does not exist, it
	// is created and the policy is set.
	SetPolicy(queue, policy string) error

	// Pause queue streaming consumer
	Pause() error

	// Start queue streaming. All queues are started by default
	Start() error

	// List all queues
	ListQueues() ([]string, error)
}

var (
	InvalidQueuePolicy = errors.New("Invalid queue policy")
)

type ClientPresence interface {
	AddHttpClient(queue string, w http.ResponseWriter) error
	AddSSEClient(queue string, rw *bufio.ReadWriter) error
	AddWSClient(queue string, ws *websocket.Conn) error
}
