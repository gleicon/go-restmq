# RestMQ

RestMQ is an HTTP based message queue. Forget protocols and alien clients.
Power up your favorite HTTP client and go.

RestMQ is implemented as a library that can be embedded in other software,
but also ships with RestMQ daemon, or `restmqd`.

## Build, run, install

Download required packages:

	make deps

Build and run dev server:

	make clean all; ./restmqd/restmqd

Install, uninstall:

	sudo make install
	sudo make uninstall

## Devops

`restmq` is the library, currently using Redis as the backend storage.
`restmqd` is the http and https server that ships with the library.

Lots of TODO in the source, examples and test cases.

## Testing

Use curl to test functionality.

Add new item *bar* into the queue *foo*:

	curl -v -d 'value=bar' http://localhost:8080/q/foo

Get the next item from queue *foo*:

	curl -v http://localhost:8080/q/foo

Use comet (SSE) to subscribe to queue *foo* and get new items as they're
created, in real time:

	curl -v http://localhost:8080/c/foo


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
