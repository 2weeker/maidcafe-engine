package tripcode

import "testing"

func TestClassicMatchesCrypt3(t *testing.T) {
	tests := map[string]string{
		"tea":                  "WokonZwxw2",
		"password":             "ozOtJW9BFA",
		"elite":                "006z00sOUM",
		"a":                    "ZnBI2EKkq.",
		"ab":                   "85qvGhCCNc",
		"#hash":                "glb/YgbeoM",
		"long passphrase here": "mOMyJIHjLI",
		"z~z":                  "DlfB7DdVcM",
		"x:y":                  "gGXDV1B60E",
		"9/9":                  "wJrVRVZ7vU",
	}
	for key, want := range tests {
		if got := Classic(key); got != want {
			t.Errorf("Classic(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestClassicProperties(t *testing.T) {
	if Classic("") != "" {
		t.Error("empty key gives no tripcode")
	}
	if Classic("12345678") != Classic("12345678ignored") {
		t.Error("only the first eight bytes of the key count, like crypt(3)")
	}
	if Classic("tea") == Classic("Tea") {
		t.Error("keys are case-sensitive")
	}
	for _, c := range Classic("anything") {
		if !(c == '.' || c == '/' || (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')) {
			t.Fatalf("unexpected character %q in tripcode", c)
		}
	}
}
