package dir

import (
	"reflect"
	"testing"

	"github.com/alx99/fly/internal/config"
	"github.com/alx99/fly/internal/state"
	"github.com/rs/zerolog"
)

func TestModel_Move(t *testing.T) {
	state := &state.State{}
	logger := zerolog.New(zerolog.TestWriter{T: t})
	settings := config.Settings{ScrollPadding: 2}
	type args struct {
		dir Direction
	}
	tests := []struct {
		name  string
		start Model
		args  args
		want  Model
	}{
		{
			name: "cursorIndex stays the same when moving up at the top",
			args: args{
				dir: Up,
			},
			start: Model{
				state:            state,
				h:                4,
				cursorIndex:      0,
				visibleFileCount: 10,
				cfg:              settings,
				logger:           logger,
			},
			want: Model{
				state:            state,
				h:                4,
				cursorIndex:      0,
				visibleFileCount: 10,
				cfg:              settings,
				logger:           logger,
			},
		},
		{
			name: "cursorIndex stays the same when moving down at the bottom",
			args: args{
				dir: Down,
			},
			start: Model{
				state:            state,
				h:                4,
				offset:           6,
				cursorIndex:      9,
				visibleFileCount: 10,
				cfg:              settings,
				logger:           logger,
			},
			want: Model{
				state:            state,
				h:                4,
				offset:           6,
				cursorIndex:      9,
				visibleFileCount: 10,
				cfg:              settings,
				logger:           logger,
			},
		},
		{
			name: "move down",
			args: args{
				dir: Down,
			},
			start: Model{
				state:            state,
				h:                4,
				cursorIndex:      0,
				visibleFileCount: 10,
				cfg:              settings,
				logger:           logger,
			},
			want: Model{
				state:            state,
				h:                4,
				cursorIndex:      1,
				visibleFileCount: 10,
				cfg:              settings,
				logger:           logger,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.start.Move(tt.args.dir)
			if !reflect.DeepEqual(got, &tt.want) {
				t.Errorf("Model.Move() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
