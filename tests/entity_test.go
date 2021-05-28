package test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type Sanitizable interface {
	Sanitize()
}

func TestSanitizable_Sanitize(t *testing.T) {

	cmpOpts := []cmp.Option{
		cmpopts.IgnoreUnexported(
			Entity1{},
			Entity2{},
			Entity3{},
			Entity4{},
			Entity5{},
			Entity6{},
			Entity7{},
		),
	}

	tests := []struct {
		name    string
		obj     Sanitizable
		want    Sanitizable
		cmpOpts []cmp.Option
	}{
		{
			name: "Test kind TEXT",
			obj: &Entity1{
				Name: "<b>name</b>",
				Text: "<p>text</p>",
				Uuid: "deadbeef",
			},
			want: &Entity1{
				Name: "name",
				Text: "<p>text</p>",
				Uuid: "deadbeef",
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test kind TEXT and HTML",
			obj: &Entity1{
				Name: "<b>name</b>",
				Text: "<iframe>text</iframe>",
				Uuid: "deadbeef",
			},
			want: &Entity1{
				Name: "name",
				Text: "",
				Uuid: "deadbeef",
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test kind TEXT and nested HTML",
			obj: &Entity1{
				Name: "<b>name</b>",
				Text: "<pre>pre<iframe>text</iframe>post</pre>",
				Uuid: "deadbeef",
			},
			want: &Entity1{
				Name: "name",
				Text: "<pre>prepost</pre>",
				Uuid: "deadbeef",
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test sanitize subfield and not disable_field",
			obj: &Entity5{
				Entity1: &Entity1{
					Name: "<b>name</b>",
					Text: "<pre>pre<iframe>text</iframe>post</pre>",
					Uuid: "deadbeef",
				},
				Entity2: &Entity2{
					Name: "<b>name</b>",
				},
			},
			want: &Entity5{
				Entity1: &Entity1{
					Name: "name",
					Text: "<pre>prepost</pre>",
					Uuid: "deadbeef",
				},
				Entity2: &Entity2{
					Name: "<b>name</b>",
				},
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test sanitize repeated message",
			obj: &Entity6{
				Entities: []*Entity1{
					{
						Name: "<b>name</b>",
						Text: "<pre>pre<iframe>text</iframe>post</pre>",
						Uuid: "deadbeef",
					},
				},
				Entity2: &Entity2{
					Name: "<b>name</b>",
				},
			},
			want: &Entity6{
				Entities: []*Entity1{
					{
						Name: "name",
						Text: "<pre>prepost</pre>",
						Uuid: "deadbeef",
					},
				},
				Entity2: &Entity2{
					Name: "<b>name</b>",
				},
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test sanitize repeated string",
			obj: &Entity2{
				Name:  "<b>name</b>",
				Value: 123,
				Uuids: []string{"<b>uuid1</b>", "<b>uuid2</b>"},
			},
			want: &Entity2{
				Name:  "name",
				Value: 123,
				Uuids: []string{"uuid1", "uuid2"},
			},
			cmpOpts: cmpOpts,
		},
		{
			name:    "Test sanitize nil message",
			obj:     &Entity7{},
			want:    &Entity7{},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test sanitize nil value in message",
			obj: &Entity1{
				Name: "<b>name</b>",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.obj.Sanitize()
			diff := cmp.Diff(tt.want, tt.obj, tt.cmpOpts...)
			if tt.want != nil && diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}

		})
	}
}
