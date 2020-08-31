
# Overview

Teffy is a library for reading & writing Google Chrome [Trace Event Format](https://docs.google.com/document/d/1CvAClvFfyA5R-PhYUmn5OOQtYMH4h6I0nSsKchNAySU/preview) files in Go.

# Quickstart

## Install

`go get github.com/omaskery/teffy`

## Reading Events

```go
package main

import (
    "fmt"
    "os"
    
    teffyio "github.com/omaskery/teffy/pkg/io"
)

func main() {
    f, _ := os.Open("some.trace")
    defer f.Close()

    trace, _ := teffyio.ParseJsonObj(f)

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
    teffyio "github.com/omaskery/teffy/pkg/io"
)

func main() {
    f, _ := os.OpenFile("some.trace", os.O_RDWR, os.ModePerm)
    defer f.Close()

    writer := teffyio.JsonArrayWriter(f)
    writer.Write(&events.BeginDuration{
        Name: "my event",
        Categories: []string{ "cool", "categories" },
        Timestamp: 0,
    })
    // your amazing code
    writer.Write(&events.EndDuration{
        Name: "my event",
        Categories: []string{ "cool", "categories" },
        Timestamp: 100,
    })
}
```

## Opinionated Event Writing Utilities

```go
package main

import (
    "github.com/omaskery/teffy/pkg/util"
)

func main() {
    t, _ := util.StartTrace("some.trace")
    defer t.Close()
    
    duration := t.BeginDuration("my event", util.WithCategories("cool", "categories"))
    defer duration.Close()
    // your amazing code
}
```
