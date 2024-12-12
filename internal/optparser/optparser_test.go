package optparser

import (
	"strconv"
	"testing"
)

func TestNum2Str(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{input: 0, expected: "0"},                     // Нулевое значение
		{input: 123, expected: "123"},                 // Положительное число
		{input: -123, expected: "-123"},               // Отрицательное число
		{input: 2147483647, expected: "2147483647"},   // Максимальное значение int32
		{input: -2147483648, expected: "-2147483648"}, // Минимальное значение int32
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.input), func(t *testing.T) {
			result := Num2Str(tt.input)
			if result != tt.expected {
				t.Errorf("Num2Str(%d) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
