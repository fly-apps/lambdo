# Lambdo

> Not quite a Lambda, almost as cool as a Lambo.

Run workloads on Fly Machines based on external events. It's like serverless, but you can run your whole code base (or whatever you want).

There are three components:

1. The server (this code base), likelly run as an app within Fly
2. Your code, which is built into a Docker image (like normal on Fly.io)
3. An SQS queue you own and populate with your own events

## The Server

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

## Your Code

You need some code that reads in a JSON string from file `/tmp/events.json`. This is an array of arbitrary events that you create via the SQS queue.

Example content of `/tmp/events.json`:

```json
[
   {"some": "object"},
   {"that": "you", "created": true}
]
```

You can either program up your own code to handle this file, or if you like the "serverless function" style, you can use a base image provided by this project.

### Use Your Code Base

One way to go about this is to make a command that can be run (`php artisan foo`, `rake foo`, `node index.js foo`, whatever)
that reads in the JSON from `/tmp/events.json` and handles each event provided.

Then, when you create an event to process, you can set that command to be run to process the events (examples on that below).

### Use A Serverless Function

There's 2 base images here you can use, which is a little more like serverless in that you can provide them a function to run for each event.

Your Docker image can use the provided images as a default for Node 20 or PHP 8.2 (or make your own), and add in your own code to the correct place.

See the [sample JS project](runtimes/js/sample-project) or the [sample PHP project](runtimes/php/sample-project) to see what that looks like.

## The SQS Queue

The SQS queue is the source of events. Sending a message to this queue will result in Machines being created to process them.

The message `Body` should be a valid JSON string (your event, its contens are arbitrary).

The message `Attributes` have up to 3 values to help the project know how to spin up a Machine and process the event.

It looks like this (forgive the lame need for escaping double quotes):

```bash
QUEUE_URL="https://sqs.<region>.amazonaws.com/<account>/<queue>"
JSON_BODY='{"foo": "bar"}'

aws sqs send-message \
  --queue-url=$QUEUE_URL \
  --message-body=$JSON_BODY \
  --message-attributes='{
  "image":{"DataType":"String","StringValue":"registry.fly.io/app:tag"},
  "size":{"DataType":"String","StringValue":"performance-2x"},
  "command":{"DataType":"String","StringValue":"[\"php\", \"artisan\", \"foo\"]"},
}'
```

There are 3 attribute values to care about:

| Attribute | Description                                                           | Default                  |
|-----------|-----------------------------------------------------------------------|--------------------------|
| `image`   | **required** - The image to run in the Machine to process that event  |                          |
| `size` | The VM size<sup>†</sup>                                               | `performance-2x`         |
| `command` | The command to run, which is the Docker `CMD` equivalent<sup>††</sup> | Your `Dockerfile`'s `CMD` |

- <sup>†</sup> Use `fly platform vm-sizes` for valid values.
- <sup>††</sup> Use the array syntax, e.g.`["foo", "--bar"]`