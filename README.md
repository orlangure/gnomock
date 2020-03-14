# Gnomock Splunk ![Build](https://github.com/orlangure/gnomock-splunk/workflows/Build/badge.svg?branch=master)

Gnomock Splunk is a [Gnomock](https://github.com/orlangure/gnomock) preset for
running tests against a real Splunk container, without mocks.

```go
package splunk_test

import (
	"fmt"
	"sort"

	"github.com/orlangure/gnomock"
	mocksplunk "github.com/orlangure/gnomock-splunk"
)

func ExampleSplunk() {
	events := []mocksplunk.Event{
		{
			Event:      "action=foo",
			Index:      "events",
			Source:     "main",
			SourceType: "http",
		},
		{
			Event:      "action=bar",
			Index:      "events",
			Source:     "main",
			SourceType: "http",
		},
	}

	p := mocksplunk.Preset(
		mocksplunk.WithLicense(true),
		mocksplunk.WithPassword("12345678"),
		mocksplunk.WithValues(events),
	)

	// created container now includes two events in "events" index
	container, err := gnomock.StartPreset(p)
	fmt.Println("error:", err)
	fmt.Println(len(container.Ports), "exposed ports:")

	defer func() {
		_ = gnomock.Stop(container)
	}()

	// Port numbers as well as container address are non-deterministic, so they
	// are skipped in this example. The usage would be:
	//
	//		container.Address("web")
	// 		container.Address("api")

	portNames := make([]string, 0)

	for portName := range container.Ports {
		portNames = append(portNames, portName)
	}

	sort.Sort(sort.StringSlice(portNames))
	fmt.Println(portNames)

	// Output:
	// error: <nil>
	// 3 exposed ports:
	// [api collector web]
}
```
