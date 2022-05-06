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
			Entity9{},
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
		{
			name: "Test entity field name do not collide with loop indexes",
			obj: &Entity9{
				Id: []string{"<b>name</b>"},
			},
			want: &Entity9{
				Id: []string{"name"},
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test entity only spaces in field name ",
			obj: &Entity2{
				Name:  "       ",
				Value: 123,
				Uuids: []string{"      ", " "},
			},
			want: &Entity2{
				Name:  "",
				Value: 123,
				Uuids: []string{"", ""},
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test entity only spaces in field name after sanitization",
			obj: &Entity2{
				Name:  " <a>       </a>",
				Value: 123,
				Uuids: []string{"<h1>      </h1> ", " <b> </b>"},
			},
			want: &Entity2{
				Name:  "",
				Value: 123,
				Uuids: []string{"", ""},
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test text and html trim",
			obj: &Entity1{
				Name: "  <b>name</b> ",
				Text: " <pre> <iframe>text</iframe> </pre> ",
				Uuid: " deadbeef ",
			},
			want: &Entity1{
				Name: "name",
				Text: "<pre>  </pre>",
				Uuid: " deadbeef ",
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test empty string after html sanitize",
			obj: &Entity1{
				Text: " <iframe>text</iframe>  ",
			},
			want: &Entity1{
				Text: "",
			},
			cmpOpts: cmpOpts,
		},
		{
			name: "Test color attribute in html field",
			obj: &Entity1{
				Text: "<h1 style=\"color:red\">text</h1>",
			},
			want: &Entity1{
				Text: "<h1 style=\"color: red\">text</h1>",
			},
			cmpOpts: cmpOpts,
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
