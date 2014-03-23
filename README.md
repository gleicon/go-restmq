# RestMQ

RestMQ is an HTTP based message queue. Forget protocols and alien clients.
Power up your favorite HTTP client and go.

RestMQ is implemented as a library that can be embedded in other software,
but also ships with RestMQ daemon, or `restmqd`.

## Build, run, install

Download required packages:

	make deps

Build and run dev server:

	make clean all
	./restmqd

Install, uninstall:

	sudo make install
	sudo make uninstall

Build Debian package:

	sudo make dpkg-deps
	make dpkg

Build RPM package:

	sudo make rpm-deps
	make rpm

## Testing

Use curl to test functionality.

Add new item *bar* into the queue *foo*:

	curl -v -d 'item=hello, world' http://localhost:8080/q/foobar

Get the next item from queue *foo*:

	curl -v http://localhost:8080/q/foobar

Use comet (SSE) to subscribe to queue *foo* and get new items as they're
created, in real time:

	(NOT WORKING YET)
	curl -v http://localhost:8080/c/foobar


## Stuff we do

## Stuff we dont do

## TODO

- add COMET and SSE
- add ws
- add pause toggle
- add policy toggle
- include https://github.com/supr/sqs
- clean up old protocols
- write leveldb abstraction
- be awesom

## Credits

See the AUTHORS and CONTRIBUTORS files for details.
