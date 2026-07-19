package poscode

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"ёжик", "ежик"},
		{"ежик", "ежик"},
		{"Ёжик", "ежик"},
		{"майка", "маика"},
		{"маика", "маика"},
		{"МАЙКА", "маика"},
		{"  Привет  ", "привет"},
		{"при вет", "привет"},
		{" Зайчик ", "заичик"},
		{"ЗАЙЧИК", "заичик"},
		{"йод", "иод"},
		{"", ""},
		{"   ", ""},
	}
	for _, c := range cases {
		if got := Normalize(c.in); got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestWordsMinLength(t *testing.T) {
	for _, w := range words {
		if n := runeLen(w); n < 5 {
			t.Errorf("word %q has rune length %d, want >= 5", w, n)
		}
	}
}

func TestWordsNoNormalizedDuplicates(t *testing.T) {
	seen := make(map[string]string, len(words))
	for _, w := range words {
		n := Normalize(w)
		if prev, ok := seen[n]; ok {
			t.Errorf("words %q and %q collide after Normalize (%q)", prev, w, n)
			continue
		}
		seen[n] = w
	}
}

func TestWordsCount(t *testing.T) {
	// Entropy sanity: the curated list should stay in the target range.
	if len(words) < 600 {
		t.Errorf("wordlist has %d words, want >= 600 for entropy", len(words))
	}
}
