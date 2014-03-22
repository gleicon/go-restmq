# Copyright 2013 restmq Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

NAME=restmq

all: deps server

deps:
	go get -v

server:
	go build -v -o restmq

clean:
	rm -f restmq

.PHONY: server
