package config

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func Test_parseNums(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"_", args{"1;2"}, []int{1, 2}},
		{"_", args{"0;5;2"}, []int{0, 5, 2}},
		{"_", args{"01;31"}, []int{1, 31}},
		{"_", args{"0;38;2;226;209;57"}, []int{0, 38, 2, 226, 209, 57}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseNums(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseNums() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseStyle(t *testing.T) {
	st := tcell.StyleDefault
	tests := []struct {
		name    string
		arg     string
		want    string
		want1   tcell.Style
		wantErr bool
	}{
		{"Badly formatted entry 1",
			"ln=31=",
			"", st, true},
		{"Badly formatted entry 2",
			"ln=01;38;5",
			"", st, true},
		{"Badly formatted entry 3",
			"ln=01;38;2;12;12",
			"", st, true},
		{"Badly formatted entry 4",
			"ln=01;48;2;12;12",
			"", st, true},
		{"Badly formatted entry 5",
			"ln=01;48;2;12;",
			"", st, true},

		{"Empty style test", "ln=0", "ln", st, false},
		{"Bold", "ln=1", "ln", st.Bold(true), false},
		{"Bold", "ln=01", "ln", st.Bold(true), false},
		{"Underline", "ln=4", "ln", st.Underline(true), false},
		{"Reverse", "ln=7", "ln", st.Reverse(true), false},
		{"Blink", "ln=5", "ln", st.Blink(true), false},
		{"Blink", "ln=6", "ln", st.Blink(true), false},
		{"Strikethrough", "ln=9", "ln", st.StrikeThrough(true), false},
		{"Italic", "ln=3", "ln", st.Italic(true), false},
		{"Dim", "ln=2", "ln", st.Dim(true), false},

		{"Background normal", "ln=41", "ln", st.Background(tcell.ColorMaroon), false},
		{"Foreground 256", "ln=38;5;1", "ln", st.Foreground(tcell.ColorMaroon), false},
		{"Foreground 256", "ln=38;5;232", "ln", st.Foreground(tcell.PaletteColor(232)), false},
		{"Foreground RGB", "ln=38;2;5;102;8", "ln", st.Foreground(tcell.NewRGBColor(5, 102, 8)), false},

		{"Foreground normal", "ln=31", "ln", st.Foreground(tcell.ColorMaroon), false},
		{"Background 256", "ln=48;5;1", "ln", st.Background(tcell.ColorMaroon), false},
		{"Background 256", "ln=48;5;232", "ln", st.Background(tcell.PaletteColor(232)), false},
		{"Background RGB", "ln=48;2;5;102;8", "ln", st.Background(tcell.NewRGBColor(5, 102, 8)), false},

		// Combos
		{"Foreground normal & bold", "ln=1;31", "ln", st.Foreground(tcell.ColorMaroon).Bold(true), false},
		{"Foreground 256 & strikethrough & dim", "ln=48;5;232;9;2", "ln", st.Background(tcell.PaletteColor(232)).StrikeThrough(true).Dim(true), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseStyle(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStyle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// If we want error, we won't bother checking
			// the style or the error
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("parseStyle() entry = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseStyle() style = %v, want %v", got1, tt.want1)
			}
		})
	}
}
