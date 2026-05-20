package sub

import "testing"

func TestSubEntitySanitize(t *testing.T) {
	e := &SubEntity{Name: " <b>hi</b> "}
	e.Sanitize()
	if e.Name != "hi" {
		t.Fatalf("Sanitize: name = %q, want %q", e.Name, "hi")
	}
}
