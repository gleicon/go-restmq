# Copyright 2010-2014 restmq authors.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

include Makefile.defs

all: server

deps:
	go get -v

forcedeps:
	go get -v -u

server:
	go build -v -o $(NAME) -ldflags "-X main.VERSION $(VERSION)"

clean:
	rm -f $(NAME)

.PHONY: server
