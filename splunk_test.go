package splunk_test

import (
	"fmt"
	"sort"
	"time"

	"github.com/orlangure/gnomock"
	mocksplunk "github.com/orlangure/gnomock-splunk"
)

func ExamplePreset() {
	events := []mocksplunk.Event{
		{
			Event:      "action=foo",
			Index:      "events",
			Source:     "app",
			SourceType: "http",
		},
		{
			Event:      "action=bar",
			Index:      "events",
			Source:     "app",
			SourceType: "http",
		},
	}

	p := mocksplunk.Preset(
		mocksplunk.WithVersion("edge"),
		mocksplunk.WithLicense(true),
		mocksplunk.WithPassword("12345678"),
		mocksplunk.WithValues(events),
		mocksplunk.WithInitTimeout(time.Second),
	)

	// created container now includes two events in "events" index
	container, err := gnomock.Start(p)
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

	sort.Strings(portNames)
	fmt.Println(portNames)

	// Output:
	// error: <nil>
	// 3 exposed ports:
	// [api collector web]
}
