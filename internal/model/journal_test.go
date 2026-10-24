package model

import "testing"

func TestJournalValidate(t *testing.T) {
	tests := []struct {
		name  string
		title string
		valid bool
	}{
		{
			name:  "judul valid",
			title: "Refleksi hari ini",
			valid: true,
		},
		{
			name:  "judul kosong",
			title: "",
			valid: false,
		},
		{
			name:  "judul hanya spasi",
			title: "   ",
			valid: false,
		},
		{
			name:  "judul lebih dari 200 karakter",
			title: string(make([]byte, 201)),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := Journal{Title: tt.title}
			err := j.Validate()

			if (err == nil) != tt.valid {
				t.Fatalf(
					"Validate() error = %v, valid expected %v",
					err,
					tt.valid,
				)
			}
		})
	}
}
