package utils

import (
	"math"
	"testing"
)

func TestGetAngle(t *testing.T) {
	tests := []struct {
		name string
		from Point
		to   Point
		want float64
	}{
		{
			name: "right",
			from: NewPoint(0, 0),
			to:   NewPoint(1, 0),
			want: 0,
		},
		{
			name: "up",
			from: NewPoint(0, 0),
			to:   NewPoint(0, -1),
			want: math.Pi / 2,
		},
		{
			name: "left",
			from: NewPoint(0, 0),
			to:   NewPoint(-1, 0),
			want: math.Pi,
		},
		{
			name: "almost left",
			from: NewPoint(0, 0),
			to:   NewPoint(-0.999999999, 0),
			want: math.Pi,
		},
		{
			name: "a bit past left",
			from: NewPoint(0, 0),
			to:   NewPoint(-1.0000001, 0),
			want: math.Pi,
		},
		{
			name: "down",
			from: NewPoint(0, 0),
			to:   NewPoint(0, 1),
			want: 3 * math.Pi / 2,
		},
		{
			name: "up right",
			from: NewPoint(0, 0),
			to:   NewPoint(1, -1),
			want: math.Pi / 4,
		},
		{
			name: "up left",
			from: NewPoint(0, 0),
			to:   NewPoint(-1, -1),
			want: 3 * math.Pi / 4,
		},
		{
			name: "down left",
			from: NewPoint(0, 0),
			to:   NewPoint(-1, 1),
			want: 5 * math.Pi / 4,
		},
		{
			name: "down right",
			from: NewPoint(0, 0),
			to:   NewPoint(1, 1),
			want: 7 * math.Pi / 4,
		},
		{
			name: "offset origin",
			from: NewPoint(10, 10),
			to:   NewPoint(11, 9),
			want: math.Pi / 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAngle(tt.from.X, tt.from.Y, tt.to.X, tt.to.Y)

			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("GetAngle(%v, %v, %v, %v) = %v; want %v",
					tt.from.X, tt.from.Y, tt.to.X, tt.to.Y, got, tt.want)
			}
		})
	}
}
