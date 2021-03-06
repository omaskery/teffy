package main

import (
	"fmt"
	"os"

	"github.com/omaskery/teffy/pkg/io"
)

func main() {
	f, err := os.Open("trace.json")
	if err != nil {
		abortWithErr("failed to open trace file", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			abortWithErr("failed to close trace file", err)
		}
	}()

	data, err := io.ParseJsonObj(f)
	if err != nil {
		abortWithErr("failed to parse trace file", err)
	}

	fmt.Printf("display time unit: %s\n", data.DisplayTimeUnit())

	if data.ControllerTraceDataKey() != "" {
		fmt.Printf("controller trace data key: %s\n", data.ControllerTraceDataKey())
	}

	if data.SystemTraceEvents() != "" {
		fmt.Println("system trace events are present")
	}

	if data.PowerTraceAsString() != "" {
		fmt.Println("power trace information is present")
	}

	stackTraceCount := 0
	for _, frame := range data.StackFrames() {
		if frame.Parent == "" {
			stackTraceCount += 1
		}
	}
	fmt.Printf("contains %v stack frames (~%v stack traces)\n", len(data.StackFrames()), stackTraceCount)

	for key, _ := range data.Metadata() {
		fmt.Printf("contains metadata '%s'\n", key)
	}

	fmt.Printf("ingested %v trace events\n", len(data.Events()))
}

func abortWithErr(reason string, err error) {
	abort(fmt.Sprintf("%s: %v", reason, err))
}

func abort(reason string) {
	_, err := os.Stderr.WriteString(reason)
	if err != nil {
		panic(fmt.Sprintf("failed while writing error to terminal: %v", err))
	}
	os.Exit(1)
}
