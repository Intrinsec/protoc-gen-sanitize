package test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEventSanitize_WKTUntouched(t *testing.T) {
	e := &Event{DetectionTime: timestamppb.New(time.Unix(123, 0))}
	e.Sanitize() // must compile (no .Sanitize() on the WKT) and not mutate it
	if e.DetectionTime == nil || e.DetectionTime.GetSeconds() != 123 {
		t.Fatalf("WKT field mutated or cleared: %v", e.DetectionTime)
	}
}
