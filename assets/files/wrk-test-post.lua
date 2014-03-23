-- wrk -c 100 -d 10s -t 2 -s test-post.lua http://localhost:8080/q/test
-- TODO: random values

wrk.method = "POST"
wrk.body   = "value=foobar"
wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
