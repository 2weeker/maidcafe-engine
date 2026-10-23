package models

import "testing"

func TestNewIdentity(t *testing.T) {
	if id := NewIdentity("nanashi", ""); !id.Anonymous() || id.Name != "" {
		t.Errorf("a name without a key is dropped: %+v", id)
	}
	id := NewIdentity("nanashi", "tea")
	if id.Name != "nanashi" || id.Trip != "WokonZwxw2" {
		t.Errorf("got %+v", id)
	}
	if hash := NewIdentity("nanashi#tea", ""); hash != id {
		t.Errorf("name#key in the name field works: %+v", hash)
	}
	if id := NewIdentity("", "tea"); id.Name != "" || id.Trip == "" {
		t.Errorf("a key alone gives a tripcode: %+v", id)
	}
}

func TestCleanName(t *testing.T) {
	tests := map[string]string{
		"  spaced   out  ":  "spaced out",
		"fake ◆trip":        "fake trip",
		"ctrl\x00\x1bchars": "ctrlchars",
		"ab★cd☆":            "abcd",
	}
	for in, want := range tests {
		if got := CleanName(in); got != want {
			t.Errorf("CleanName(%q) = %q, want %q", in, got, want)
		}
	}
	if got := CleanName("ああああああああああああああああああああああああああああああああああ"); len([]rune(got)) != MaxName {
		t.Errorf("name is cut to %d runes, got %d", MaxName, len([]rune(got)))
	}
}
