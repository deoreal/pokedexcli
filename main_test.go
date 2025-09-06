package main

import (
	"fmt"
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			input:    "   ",
			expected: []string{},
		},
		{
			input:    "",
			expected: []string{},
		},
		{
			input:    "single",
			expected: []string{"single"},
		},
		{
			input:    "  MULTIPLE   SPACES   BETWEEN   WORDS  ",
			expected: []string{"multiple", "spaces", "between", "words"},
		},
		{
			input:    "\t\nTabsAndNewlines\t\n",
			expected: []string{"tabsandnewlines"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		fmt.Println("actual", actual)
		fmt.Println("expectedWord", c.expected)
		if len(actual) != len(c.expected) {
			t.Errorf("cleanInput(%q) returned slice of length %d, expected length %d",
				c.input, len(actual), len(c.expected))
			continue
		}

		// Check each word in the slice
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("cleanInput(%q) returned word %q at index %d, expected %q",
					c.input, word, i, expectedWord)
			}
		}
	}
}
