package model

import (
	"testing"
)

func TestAsset1_Sanitize(t *testing.T) {
	type fields struct {
		Name string
		Text string
		Uuid string
	}
	tests := []struct {
		name string
		init fields
		want fields
	}{
		{
			name: "test00",
			init: fields{"<b>name</b>", "<p>text</p>", "deadbeef"},
			want: fields{"name", "<p>text</p>", "deadbeef"},
		},
		{
			name: "test01",
			init: fields{"<b>name</b>", "<iframe>text</iframe>", "deadbeef"},
			want: fields{"name", "", "deadbeef"},
		},
		{
			name: "test02",
			init: fields{"<b>name</b>", "<pre>pre<iframe>text</iframe>post</pre>", "deadbeef"},
			want: fields{"name", "<pre>prepost</pre>", "deadbeef"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Asset1{
				Name: tt.init.Name,
				Text: tt.init.Text,
				Uuid: tt.init.Uuid,
			}
			m.Sanitize()

			if m.Name != tt.want.Name {
				t.Errorf("Asset1.Sanitize() Name = `%v`, wanted `%v`", m.Name, tt.want.Name)
			}
			if m.Text != tt.want.Text {
				t.Errorf("Asset1.Sanitize() Text = `%v`, wanted `%v`", m.Text, tt.want.Text)
			}
			if m.Uuid != tt.want.Uuid {
				t.Errorf("Asset1.Sanitize() Uuid = `%v`, wanted `%v`", m.Uuid, tt.want.Uuid)
			}
		})
	}
}
