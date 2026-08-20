package config

import "testing"

func TestCheckBindAddress(t *testing.T) {
	cases := []struct {
		host     string
		override bool
		wantErr  bool
	}{
		{"127.0.0.1", false, false},    // loopback
		{"::1", false, false},          // loopback v6
		{"localhost", false, false},    // hostname loopback
		{"192.168.1.10", false, false}, // RFC1918
		{"10.0.0.5", false, false},     // RFC1918
		{"100.100.0.1", false, false},  // CGNAT mesh
		{"8.8.8.8", false, true},       // public → refused
		{"0.0.0.0", false, true},       // unspecified → refused
		{"example.com", false, true},   // unverifiable hostname → refused
		{"8.8.8.8", true, false},       // override permits public
	}
	for _, c := range cases {
		err := CheckBindAddress(c.host, c.override)
		if (err != nil) != c.wantErr {
			t.Errorf("CheckBindAddress(%q, override=%v) err=%v, wantErr=%v", c.host, c.override, err, c.wantErr)
		}
	}
}
