# Copyright 2009-2013 The RestMQ Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PREFIX=/opt/restmq

all: restmqd/restmqd

deps:
	go get -u github.com/fiorix/go-redis/redis
	go get -u github.com/fiorix/go-web/http
	go get -u github.com/fiorix/go-web/remux

restmqd/restmqd:
	(cd restmqd; go build)

clean:
	rm -f restmqd/restmqd

install: restmqd/restmqd
	mkdir -p ${PREFIX}
	install -o root -m 0755 restmqd/restmqd ${PREFIX}
	install -o root -m 0640 restmqd.xml ${PREFIX}
	mkdir -m 0750 -p ${PREFIX}/cert
	install -o root -m 0755 cert/mkcert.sh ${PREFIX}/cert

uninstall:
	rm -rf ${PREFIX}
