# Copyright 2013 restmq Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

PREFIX=$(DESTDIR)/opt/restmq

all: server

deps:
	make -C src deps

forcedeps:
	make -C src forcedeps

server:
	make -C src
	@cp src/restmqd .

clean:
	make -C src clean
	@rm -f ./restmqd

install: server
	install -m 750 -d ${PREFIX}
	install -m 750 restmqd ${PREFIX}
	install -m 640 restmq.conf ${PREFIX}
	install -m 750 -d ${PREFIX}/ssl
	install -m 750 ssl/Makefile ${PREFIX}/ssl
	install -m 750 -d ${PREFIX}/assets
	rsync -rupE assets ${PREFIX}
	find ${PREFIX}/assets -type f -exec chmod 640 {} \;
	find ${PREFIX}/assets -type d -exec chmod 750 {} \;
	#chown -R www-data: ${PREFIX}

uninstall:
	rm -rf ${PREFIX}

dpkg-deps:
	apt-get install build-essential debhelper devscripts dh-make fakeroot lintian

make dpkg:
	dpkg-buildpackage -b -uc -rfakeroot

.PHONY: server
