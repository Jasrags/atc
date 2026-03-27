package heading

import "testing"

func TestDelta(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{0, 90, 90},
		{90, 0, -90},
		{350, 10, 20},
		{10, 350, -20},
		{0, 180, -180}, // 180 is ambiguous; formula yields -180
		{180, 0, -180},
	}

	for _, tt := range tests {
		got := Delta(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Delta(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestAbsDelta(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{0, 90, 90},
		{90, 0, 90},
		{350, 10, 20},
		{10, 350, 20},
		{0, 180, 180},
		{180, 0, 180},
		{0, 0, 0},
	}

	for _, tt := range tests {
		got := AbsDelta(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("AbsDelta(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
