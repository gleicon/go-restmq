// Copyright 2010-2014 restmq authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// RestMQ protocol in Go, based on
// https://github.com/gleicon/restmq/blob/master/src/restmq/core.py

package restmq

import (
	"io"
)

type ClientPresence struct {
	presence            map[string][]io.Writer
	lastPresenceUsedIdx map[string]int
	queuesWatchdog      map[string]chan string
}

func NewClientPresence() *ClientPresence {
	cp := ClientPresence{}
	cp.presence = make(map[string][]io.Writer)
	cp.lastPresenceUsedIdx = make(map[string]int)
	cp.queuesWatchdog = make(map[string]chan string)
	return &cp
}

func (cp *ClientPresence) dispatcher(queue string, watchdog *chan string) {

}

func (cp *ClientPresence) supervisor(queues []string) {
	for _, q := range queues {
		qch := make(chan string)
		cp.queuesWatchdog[q] = qch
		go cp.dispatcher(q, &qch)
	}
}

func (cp *ClientPresence) Add(queue string, w io.Writer) error {
	cp.presence[queue] = append(cp.presence[queue], w)
	return nil
}

func (cp *ClientPresence) Broadcast(queue string, message []byte) error {
	for i, w := range cp.presence[queue] {
		if w != nil {
			_, err := w.Write(message)
			if err != nil {
				cp.presence[queue][i] = nil
			}
		} else {
			cp.presence[queue] = append(cp.presence[queue][:i], cp.presence[queue][i+1:]...)
		}
	}
	return nil
}

func (cp *ClientPresence) RoundRobin(queue string, message []byte) error {
	if cp.lastPresenceUsedIdx[queue] > len(cp.presence[queue]) {
		cp.lastPresenceUsedIdx[queue] = 0
	}
	i := cp.lastPresenceUsedIdx[queue]
	w := cp.presence[queue][i]
	cp.lastPresenceUsedIdx[queue]++
	if w != nil {
		_, err := w.Write(message)
		if err != nil {
			cp.presence[queue][i] = nil
		}
	} else {
		cp.presence[queue] = append(cp.presence[queue][:i], cp.presence[queue][i+1:]...)
	}
	return nil
}

func (cp *ClientPresence) Members(queue string) []io.Writer {
	return cp.presence[queue]
}
