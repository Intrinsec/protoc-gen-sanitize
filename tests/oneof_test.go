package test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestObservableSanitize(t *testing.T) {
	// message member: dispatched to Account.Sanitize()
	o := &Observable{Value: &Observable_Account{Account: &Account{Name: " <b>acct</b> "}}}
	o.Sanitize()
	if got := o.GetAccount().GetName(); got != "acct" {
		t.Fatalf("account name = %q, want %q", got, "acct")
	}

	// string member: sanitized in place on the wrapper
	o2 := &Observable{Value: &Observable_Raw{Raw: " <i>raw</i> "}}
	o2.Sanitize()
	if got := o2.GetRaw(); got != "raw" {
		t.Fatalf("raw = %q, want %q", got, "raw")
	}

	// WKT member: left untouched, no .Sanitize() emitted, no panic
	o3 := &Observable{Value: &Observable_SeenAt{SeenAt: timestamppb.New(time.Unix(7, 0))}}
	o3.Sanitize()
	if o3.GetSeenAt().GetSeconds() != 7 {
		t.Fatalf("seen_at mutated: %v", o3.GetSeenAt())
	}

	// nil oneof: safe
	(&Observable{}).Sanitize()
}
