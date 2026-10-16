package models

import "testing"

func TestChannelNames(t *testing.T) {
	for in, want := range map[string]string{"#Tea-Room": "tea-room", "  gamers ": "gamers", "#a": "a"} {
		if got := NormalizeChannel(in); got != want {
			t.Errorf("NormalizeChannel(%q) = %q, want %q", in, got, want)
		}
	}
	for name, ok := range map[string]bool{
		"tea": true, "tea-room": true, "a": true, "late_night2": true,
		"": false, "-tea": false, "_tea": false, "tea room": false, "café": false,
		"abcdefghijklmnopqrstuvwxy": false, "abcdefghijklmnopqrstuvwx": true,
	} {
		if ValidChannel(name) != ok {
			t.Errorf("ValidChannel(%q) = %v, want %v", name, !ok, ok)
		}
	}
}
