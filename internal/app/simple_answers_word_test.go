package app

import "testing"

func TestSimpleArithmetic_WordForms(t *testing.T) {
	cases := map[string]string{
		"what is 17 times 23":       "391.",
		"what is 2 plus 2":          "4.",
		"what is 10 minus 4":        "6.",
		"what is 20 divided by 5":   "4.",
		"what is 6 multiplied by 7": "42.",
		"what's 8 x 9":              "72.",
		"what is 100 / 4":           "25.",
	}
	for raw, want := range cases {
		got, ok := simpleArithmeticAnswer(raw)
		if !ok || got != want {
			t.Fatalf("%q → (%q, %v), want (%q, true)", raw, got, ok, want)
		}
	}
	// Non-arithmetic must stay unhandled so it routes normally.
	if _, ok := simpleArithmeticAnswer("what times should we meet"); ok {
		t.Fatal("prose containing 'times' must not be treated as arithmetic")
	}
}
