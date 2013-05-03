# RestMQ

RestMQ is an HTTP based message queue. Forget protocols and alien clients.
Power up your favorite HTTP client and go.

RestMQ is implemented as a library that can be embedded in other software,
but also ships with RestMQ daemon, or `restmqd`.

## Requirements

Make sure the Go installation is ok, $GOPATH is set, and use the following
command to install required packages to build RestMQ:

	go get github.com/fiorix/go-redis/redis
	go get github.com/fiorix/go-web/http
	go get github.com/fiorix/go-web/remux

## Building and running

Both `restmq` and `restmqd` can be built with `go build`.

Build and execute RestMQ daemon:

	cd restmqd
	go build; ./restmqd --config=../restmqd.xml

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
