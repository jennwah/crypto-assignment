package wallet_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/jennwah/crypto-assignment/internal/domain/wallet"
)

func TestConvertFromCentsToDollarsString(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{
			name:     "Zero cents",
			input:    0,
			expected: "0.00",
		},
		{
			name:     "One cent",
			input:    1,
			expected: "0.01",
		},
		{
			name:     "Ten cents",
			input:    10,
			expected: "0.10",
		},
		{
			name:     "One dollar",
			input:    100,
			expected: "1.00",
		},
		{
			name:     "One dollar and ninety-nine cents",
			input:    199,
			expected: "1.99",
		},
		{
			name:     "Ten thousand cents",
			input:    10000,
			expected: "100.00",
		},
		{
			name:     "Max uint64",
			input:    ^uint64(0), // 18446744073709551615
			expected: decimal.NewFromUint64(^uint64(0)).Div(decimal.NewFromInt(100)).StringFixed(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wallet.ConvertFromCentsToDollarsString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
