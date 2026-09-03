package logging_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/simorgh3196/golang-microservice-sample/pkg/logging"
)

func TestMaskedString_LogValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    logging.MaskedString
		expected string
	}{
		{
			name:     "空文字の場合",
			input:    logging.MaskedString(""),
			expected: "",
		},
		{
			name:     "8文字以下の場合、全体がマスクされる",
			input:    logging.MaskedString("12345678"),
			expected: "****",
		},
		{
			name:     "9文字以上の場合、先頭4文字と末尾3文字以外がマスクされる",
			input:    logging.MaskedString("123456789"),
			expected: "1234****789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := tt.input.LogValue().String()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
