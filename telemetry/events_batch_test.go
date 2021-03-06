// Copyright 2019 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/newrelic/newrelic-telemetry-sdk-go/internal"
)

func testEventBatchJSON(t testing.TB, batch *eventBatch, expect string) {
	if th, ok := t.(interface{ Helper() }); ok {
		th.Helper()
	}
	reqs, err := newRequests(batch, "apiKey", defaultSpanURL, "userAgent")
	if nil != err {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatal(reqs)
	}
	req := reqs[0]
	actual := string(req.UncompressedBody)
	compact := compactJSONString(expect)
	if actual != compact {
		t.Errorf("\nexpect=%s\nactual=%s\n", compact, actual)
	}

	body, err := ioutil.ReadAll(req.Request.Body)
	req.Request.Body.Close()
	if err != nil {
		t.Fatal("unable to read body", err)
	}
	if len(body) != req.compressedBodyLength {
		t.Error("compressed body length mismatch",
			len(body), req.compressedBodyLength)
	}
	uncompressed, err := internal.Uncompress(body)
	if err != nil {
		t.Fatal("unable to uncompress body", err)
	}
	if string(uncompressed) != string(req.UncompressedBody) {
		t.Error("request JSON mismatch", string(uncompressed), string(req.UncompressedBody))
	}
}

func TestEventsPayloadSplit(t *testing.T) {
	t.Parallel()

	// test len 0
	ev := &eventBatch{}
	split := ev.split()
	if split != nil {
		t.Error(split)
	}

	// test len 1
	ev = &eventBatch{Events: []Event{{EventType: "a"}}}
	split = ev.split()
	if split != nil {
		t.Error(split)
	}

	// test len 2
	ev = &eventBatch{Events: []Event{{EventType: "a"}, {EventType: "b"}}}
	split = ev.split()
	if len(split) != 2 {
		t.Error("split into incorrect number of slices", len(split))
	}

	testEventBatchJSON(t, split[0].(*eventBatch), `[{"eventType":"a","timestamp":-6795364578871}]`)
	testEventBatchJSON(t, split[1].(*eventBatch), `[{"eventType":"b","timestamp":-6795364578871}]`)

	// test len 3
	ev = &eventBatch{Events: []Event{{EventType: "a"}, {EventType: "b"}, {EventType: "c"}}}
	split = ev.split()
	if len(split) != 2 {
		t.Error("split into incorrect number of slices", len(split))
	}
	testEventBatchJSON(t, split[0].(*eventBatch), `[{"eventType":"a","timestamp":-6795364578871}]`)
	testEventBatchJSON(t, split[1].(*eventBatch), `[{"eventType":"b","timestamp":-6795364578871},{"eventType":"c","timestamp":-6795364578871}]`)
}

func TestEventsJSON(t *testing.T) {
	t.Parallel()

	batch := &eventBatch{Events: []Event{
		{}, // Empty
		{ // with everything
			EventType:  "testEvent",
			Timestamp:  time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
			Attributes: map[string]interface{}{"zip": "zap"},
		},
	}}

	testEventBatchJSON(t, batch, `[
		{
		  "eventType":"",
			"timestamp":-6795364578871
		},
		{
			"eventType":"testEvent",
			"timestamp":1417136460000,
			"zip":"zap"
		}
	]`)
}
