[![PkgGoDev](https://pkg.go.dev/badge/github.com/omaskery/teffy)](https://pkg.go.dev/github.com/omaskery/teffy)

# Overview

Teffy is a library for reading & writing Google Chrome [Trace Event Format](https://docs.google.com/document/d/1CvAClvFfyA5R-PhYUmn5OOQtYMH4h6I0nSsKchNAySU/preview) files in Go.

# Quickstart

## Install

`go get github.com/omaskery/teffy`

The package is split into three main parts:
 * `events` - the logical representation of trace events
 * `io` - the ability to read/write events to files (including streaming)
 * `utils/trace` - opinionated utilities for generating traces

## Reading Events

```go
package main

import (
    "fmt"
    "os"
    
    tio "github.com/omaskery/teffy/pkg/io"
)

func main() {
    f, _ := os.Open("some.trace")
    defer f.Close()

    trace, _ := tio.ParseJsonObj(f)

    for _, event := range trace.Events() {
        fmt.Printf("- %v\n", event)
    }
}
```

## Writing Events

```go
package main

import (
    "os"
    
    "github.com/omaskery/teffy/pkg/events"
    tio "github.com/omaskery/teffy/pkg/io"
)

func main() {
    f, _ := os.OpenFile("some.trace", os.O_RDWR, os.ModePerm)
    defer f.Close()

    data := tio.TefData{}
    data.Write(&events.BeginDuration{
        Name: "my event",
        Categories: []string{ "cool", "categories" },
        Timestamp: 0,
    })
    // your amazing code
    data.Write(&events.EndDuration{
        Name: "my event",
        Categories: []string{ "cool", "categories" },
        Timestamp: 100,
    })

    _ = tio.WriteJsonObject(f, data)
}
```

## Opinionated Event Writing Utilities

```go
package main

import (
    "github.com/omaskery/teffy/pkg/util/trace"
)

func main() {
    t, _ := trace.TraceToFile("some.trace")
    defer t.Close()
    
    defer t.BeginDuration("my event", trace.WithCategories("cool", "categories")).End()

    // your amazing code
    t.Instant("wow a thing happened!", trace.WithStackTrace())
}
```
