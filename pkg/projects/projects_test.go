package projects

import "testing"

func TestProjects_NameExists(t *testing.T) {
	for _, tt := range []struct {
		name string
		want bool
	}{
		{
			name: "00-mighty-morphin-power-ranger",
			want: false,
		},
		{
			name: "01-snake-eater",
			want: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := Exists(tt.name); got != tt.want {
				t.Errorf("NameExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
