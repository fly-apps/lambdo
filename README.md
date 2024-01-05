# Lambdo

> Not quite a Lambda, almost as cool as a Lambo.

Run workloads on Fly Machines based on external events. It's like AWS Lambda, but Flyier.

There are three components:

1. The server (this code base), run as an app within Fly
2. Your code, which runs off of a Lambdo base image on Fly.io
   - It provides some event handling code, similar to how you create a Lambda function in AWS
3. An SQS queue you own and populate with your own events

## Server

Compile the program:

```bash
# For current OS:
go build -o bin/lambdo

# For Fly VMs:
GOOS=linux GOARCH=amd64 go build
```

Run the program, with some environment variables:

```bash
# Any valid AWS credential env vars will do
AWS_REGION=us-east-2 \
AWS_PROFILE=some-profile \
LAMBDO_FLY_TOKEN="$(fly auth token)" \
LAMBDO_FLY_APP=some-app \
LAMBDO_FLY_REGION=bos \
LAMBDO_SQS_QUEUE_URL
=https://sqs.<region>.amazonaws.com/<account>/<queue> \
bin/lambdo
```

## Run Your Code

You'll want to make a `Dockerfile` that uses a Lambdo base image, and then add in your own code. See the [sample JS project](runtimes/js/sample-project).

When you send images into your SQS queue, you need to provide the base image to run, e.g.:

```bash
QUEUE_URL="https://sqs.<region>.amazonaws.com/<account>/<queue>"
JSON_BODY='{"foo": "bar"}'
IMAGE="registry.fly.io/some-app:some-tag"

 aws sqs send-message \
  --queue-url=$QUEUE_URL \
  --message-body=$JSON_BODY \
  --message-attributes='{"image":{"DataType":"String","StringValue":"$IMAGE"}}'
```
