# Copyright 2010-2014 restmq authors.  All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

TARGET=$(DESTDIR)/opt/restmq
include src/Makefile.defs

SRPM_PKG=restmq-$(VERSION)
SRPM_TGZ=$(SRPM_PKG).tar.gz


all: server

deps:
	make -C src deps

forcedeps:
	make -C src forcedeps

server:
	VERSION=$(VERSION) make -C src
	@cp src/restmqd .

clean:
	make -C src clean
	@rm -f ./restmqd

install: server
	install -m 750 -d $(TARGET)
	install -m 750 restmqd $(TARGET)
	install -m 640 restmq.conf $(TARGET)
	install -m 750 -d $(TARGET)/ssl
	install -m 640 ssl/Makefile $(TARGET)/ssl
	install -m 750 -d $(TARGET)/assets
	rsync -rupE assets $(TARGET)
	find $(TARGET)/assets -type f -exec chmod 640 {} \;
	find $(TARGET)/assets -type d -exec chmod 750 {} \;
	#chown -R www-data: $(TARGET)

uninstall:
	rm -rf $(TARGET)

dpkg-deps:
	apt-get install build-essential git mercurial debhelper devscripts dh-make fakeroot lintian

make dpkg:
	dpkg-buildpackage -b -uc -rfakeroot

rpm-deps:
	yum install gcc make git mercurial rpm-build redhat-rpm-config

rpm:
	mkdir -p $(HOME)/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	echo '%_topdir $(HOME)/rpmbuild' > ~/.rpmmacros
	git archive --format tar --prefix=$(SRPM_PKG)/ HEAD . | gzip > $(HOME)/rpmbuild/SOURCES/$(SRPM_TGZ)
	cp redhat/restmq.spec $(HOME)/rpmbuild/SPECS
	rpmbuild -ba $(HOME)/rpmbuild/SPECS/restmq.spec

.PHONY: server
