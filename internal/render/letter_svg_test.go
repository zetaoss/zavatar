// internal/render/letter_svg_test.go
package render

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_pickLetters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty name",
			input: "",
			want:  "?",
		},
		{
			name:  "whitespace only",
			input: "   \t\n ",
			want:  "?",
		},
		{
			name:  "single hangul word",
			input: "홍길동",
			want:  "홍",
		},
		{
			name:  "hangul with space -> two letters",
			input: "홍 길동",
			want:  "홍길",
		},
		{
			name:  "hangul multiple words -> max two letters",
			input: "홍 길 동",
			want:  "홍길",
		},
		{
			name:  "emoji and hangul",
			input: "🙂 꺽정",
			want:  "꺽",
		},
		{
			name:  "english single word",
			input: "alice",
			want:  "A",
		},
		{
			name:  "english two words -> initials",
			input: "alice bob",
			want:  "AB",
		},
		{
			name:  "mixed english casing",
			input: "aLiCe bOB",
			want:  "AB",
		},
		{
			name:  "numbers and english",
			input: "123 abc",
			want:  "1A",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, pickLetters(tc.input))
		})
	}
}
