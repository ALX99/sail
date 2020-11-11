package config

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestMatchCommand(t *testing.T) {
	type args struct {
		k Key
		m map[Key]KeyBinding
	}
	tests := []struct {
		name  string
		args  args
		want  Command
		want1 map[Key]KeyBinding
	}{
		{
			"Check MoveUp 1",
			args{"e", nil},
			MoveUp,
			nil,
		},
		{
			"Check MoveUp 2",
			args{keyToKey[tcell.KeyUp], nil},
			MoveUp,
			nil,
		},
		{
			"Check invalid binding",
			args{"[", nil},
			Nil,
			nil,
		},
		{
			"Check doublekey binding 1",
			args{"g", nil},
			Nil,
			map[Key]KeyBinding{"g": {MoveTop, nil}},
		},
		{
			"Check doublekey binding 2",
			args{"g", map[Key]KeyBinding{"g": {MoveTop, nil}}},
			MoveTop,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MatchCommand(tt.args.k, tt.args.m)
			if got != tt.want {
				t.Errorf("MatchCommand() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MatchCommand() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
