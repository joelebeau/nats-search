# NATS Search

A simple stream searching tool for NATS JetStream.

I find myself searching for messages in NATS JetStream somewhat often, and I've also had colleagues ask insist "there MUST be a way to search through streams with the CLI, right???" - at least the last time I checked there is not.

So here we are, building a custom stream viewer. The idea is just to be able to print out the elements of a stream, optionally with a filter. I would like to add some of the same options as `nats stream view` as well as this matures.

## Installation
I may work on pre-built releases in the future, but for now the best way to install is to use `go install`:

```bash
go install github.com/joelebeau/nats-search@latest
```
