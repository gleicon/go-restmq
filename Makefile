# Copyright 2013 restmq Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PREFIX=/opt/restmq

all: server

deps:
	make -C src deps

forcedeps:
	make -C src forcedeps

server:
	make -C src
	@cp src/restmq ./restmqd

clean:
	make -C src clean
	@rm -f ./restmqd

install: server
	mkdir -m 750 -p ${PREFIX}
	install -m 750 restmqd ${PREFIX}
	install -m 640 restmq.conf ${PREFIX}
	mkdir -m 750 -p ${PREFIX}/SSL
	install -m 750 ssl/Makefile ${PREFIX}/SSL
	mkdir -m 750 -p ${PREFIX}/assets
	rsync -rupE assets ${PREFIX}
	find ${PREFIX}/assets -type f -exec chmod 640 {} \;
	find ${PREFIX}/assets -type d -exec chmod 750 {} \;
	#chown -R www-data: ${PREFIX}

uninstall:
	rm -rf ${PREFIX}

.PHONY: server
