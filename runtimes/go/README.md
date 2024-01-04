# TODO

A Go runtime requires a bit more work, we can steal ideas from Lambda [[1](https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html)], [[2](https://github.com/aws/aws-lambda-go/blob/main/lambda/entry.go)].

1. A running web server in the container that will return the JSON events
2. A package that the user-code incorporates, which takes the handler function, calls the localhost HTTP server, gets the events, and calls the handler function