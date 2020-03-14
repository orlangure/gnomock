package splunk_test

import (
	"fmt"

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
	fmt.Println(len(container.Ports), err)

	defer func() {
		_ = gnomock.Stop(container)
	}()

	// Output:
	// 3 <nil>
}
