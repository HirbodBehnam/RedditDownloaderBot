package helpers

import "testing"

func TestGetRedGifsID(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				url: "https://redgifs.com/watch/happyrequiredbellsnake",
			},
			want: "happyrequiredbellsnake",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRedGifsID(tt.args.url); got != tt.want {
				t.Errorf("GetRedGifsID() = %v, want %v", got, tt.want)
			}
		})
	}
}
