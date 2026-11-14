package proc

import "testing"

func TestParseStatPPID(t *testing.T) {
	tests := []struct {
		stat    string
		want    int
		wantErr bool
	}{
		{stat: "1234 (bash) S 1200 1234 1234 34816", want: 1200},
		{stat: "99 (my (odd) app) R 42 99 99 0", want: 42},
		{stat: "garbage", wantErr: true},
		{stat: "5 (x)", wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseStatPPID(tt.stat)
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("parseStatPPID(%q) = %d, %v; want %d, err %v", tt.stat, got, err, tt.want, tt.wantErr)
		}
	}
}
