# go-buffer

go-buffer represents a Generic buffer that asynchronously flushes its contents.

## Motivation

Provide generic buffer for user, which is useful for applications that need to aggregate data / write to an external storage before flush it.
Compatible with various data types (Container), providing error handling and secure Close.

## Features

- Periodic automatic flush
- Manually flush
- Safely Close
- Error Channel for error handling
- Generic Support

## Install

```shell
go get github.com/Kevinello/go-buffer@latest
```

### ⚠️ Retracted versions

Versions released **erroneously**, Please do not install these versions([retracted versions](https://go.dev/ref/mod#go-mod-file-retract))

- v0.0.1
- v0.0.2
