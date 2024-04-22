# go-temporal-signal-poc
A proof-of-concept for using Temporal signals in Go

## How to run

1. Start a Temporal server:
```shell
temporal server start-dev
```

2. Install dependencies and start worker:
```shell
go mod vendor
go run worker/main.go
```

3. Trigger a workflow:
```shell
go run dag/main.go
```
