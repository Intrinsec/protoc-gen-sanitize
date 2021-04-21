package test

import (
	"reflect"
	"testing"
)

type Sanitizable interface {
	Sanitize() error
}

func TestSanitizable_Sanitize(t *testing.T) {
	tests := []struct {
		name string
		obj  Sanitizable
		want Sanitizable
	}{
		{
			name: "Test kind TEXT",
			obj: &Asset1{
				Name: "<b>name</b>",
				Text: "<p>text</p>",
				Uuid: "deadbeef",
			},
			want: &Asset1{
				Name: "name",
				Text: "<p>text</p>",
				Uuid: "deadbeef",
			},
		},
		{
			name: "Test kind TEXT and HTML",
			obj: &Asset1{
				Name: "<b>name</b>",
				Text: "<iframe>text</iframe>",
				Uuid: "deadbeef",
			},
			want: &Asset1{
				Name: "name",
				Text: "",
				Uuid: "deadbeef",
			},
		},
		{
			name: "Test kind TEXT and nested HTML",
			obj: &Asset1{
				Name: "<b>name</b>",
				Text: "<pre>pre<iframe>text</iframe>post</pre>",
				Uuid: "deadbeef",
			},
			want: &Asset1{
				Name: "name",
				Text: "<pre>prepost</pre>",
				Uuid: "deadbeef",
			},
		},
		{
			name: "Test sanitize subfield and not disable_field",
			obj: &Asset5{
				Asset1: &Asset1{
					Name: "<b>name</b>",
					Text: "<pre>pre<iframe>text</iframe>post</pre>",
					Uuid: "deadbeef",
				},
				Asset2: &Asset2{
					Name: "<b>name</b>",
				},
			},
			want: &Asset5{
				Asset1: &Asset1{
					Name: "name",
					Text: "<pre>prepost</pre>",
					Uuid: "deadbeef",
				},
				Asset2: &Asset2{
					Name: "<b>name</b>",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.obj.Sanitize()

			if !reflect.DeepEqual(tt.obj, tt.want) {
				t.Errorf("Sanitize() got: `%+v`, wanted: `%+v`", tt.obj, tt.want)
			}
		})
	}
}
