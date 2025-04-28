package qrcode

import (
	"testing"

	"github.com/i9si-sistemas/assert"
)

func TestSymbolBasic(t *testing.T) {
	size := 10
	quietZoneSize := 4

	m := newSymbol(size, quietZoneSize)
	assert.Equal(t, m.size, size+quietZoneSize*2)

	for i := range size {
		for j := range size {
			v := m.get(i, j)
			assert.False(t, v)
			assert.True(t, m.empty(i, j))

			value := i*j%2 == 0
			m.set(i, j, value)

			v = m.get(i, j)
			assert.Equal(t, v, value)
			assert.False(t, m.empty(i, j), "symbol ignores set bits")
		}
	}
}

func TestSymbolPenalties(t *testing.T) {
	tests := []struct {
		pattern          [][]bool
		expectedPenalty1 int
		expectedPenalty2 int
		expectedPenalty3 int
		expectedPenalty4 int
	}{
		{
			[][]bool{
				{white, black, white, black, white, black},
				{black, white, black, white, black, white},
				{white, black, white, black, white, black},
				{black, white, black, white, black, white},
				{white, black, white, black, white, black},
				{black, white, black, white, black, white},
			},
			0, 
			0, 
			0, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, black, white, black},
				{black, white, black, white, black, white},
				{white, black, white, black, white, black},
				{black, white, black, white, black, white},
				{white, black, white, black, white, black},
				{black, white, black, white, black, white},
			},
			0, 
			0, 
			0, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, white, white, white},
				{black, white, black, white, black, white},
				{white, black, white, black, white, black},
				{black, white, black, white, black, white},
				{white, black, white, black, white, black},
				{black, white, black, white, black, white},
			},
			4, 
			0, 
			0, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, white, white, white, white},
				{black, white, black, white, black, white, black},
				{black, white, white, white, white, white, black},
				{black, white, black, white, black, white, black},
				{black, white, white, white, white, white, black},
				{black, white, black, white, black, white, black},
				{black, white, white, white, white, white, white},
			},
			28,
			0, 
			0, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, black, white, black},
				{white, white, black, white, black, white},
				{white, black, white, black, white, black},
				{black, white, black, black, black, white},
				{white, black, black, black, white, black},
				{black, white, black, white, black, white},
			},
			-1,
			6, 
			0, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, white, white, black},
				{white, white, white, white, white, black},
				{white, white, white, white, white, black},
				{white, white, white, white, white, black},
				{white, white, white, white, white, black},
				{white, white, white, white, white, black},
			},
			-1,
			60,
			0, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, white, white, black},
				{white, white, white, white, white, black},
				{black, black, white, black, white, black},
				{black, black, white, black, white, black},
				{black, black, white, black, white, black},
				{black, black, white, black, white, black},
			},
			-1,
			21,
			0, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
				{white, white, white, white, black, white, black, black, black, white, black, white},
			},
			-1,
			-1,
			480,
			-1,
		},
		{
			[][]bool{
				{black, white, white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white, white, white},
				{black, white, white, white, white, white, white, white, white, white, white, white},
				{black, white, white, white, white, white, white, white, white, white, white, white},
				{black, black, white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white, white, white},
				{black, black, white, white, white, white, white, white, white, white, white, white},
				{white, black, white, white, white, white, white, white, white, white, white, white},
				{white, black, white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white, white, white},
				{white, black, white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white, white, white},
			},
			-1,
			-1,
			80, 
			-1,
		},
		{
			[][]bool{
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
			},
			-1,
			-1,
			-1,
			100,
		},
		{
			[][]bool{
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
			},
			-1,
			-1,
			-1,
			100, 
		},
		{
			[][]bool{
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
			},
			-1,
			-1,
			-1,
			0,
		},
		{
			[][]bool{
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
			},
			-1,
			-1,
			-1,
			20,
		},
		{
			[][]bool{
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
			},
			-1,
			-1,
			-1,
			30,
		},
		{
			[][]bool{
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, white},
				{white, white, white, white, white, white, white, white, white, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
				{black, black, black, black, black, black, black, black, black, black},
			},
			-1,
			-1,
			-1,
			30,
		},
	}

	for _, test := range tests {
		s := newSymbol(len(test.pattern[0]), 4)
		s.set2dPattern(0, 0, test.pattern)

		penalty1 := s.penalty1()
		penalty2 := s.penalty2()
		penalty3 := s.penalty3()
		penalty4 := s.penalty4()

		ok := true

		if test.expectedPenalty1 != -1 && test.expectedPenalty1 != penalty1 {
			ok = false
		}
		if test.expectedPenalty2 != -1 && test.expectedPenalty2 != penalty2 {
			ok = false
		}
		if test.expectedPenalty3 != -1 && test.expectedPenalty3 != penalty3 {
			ok = false
		}
		if test.expectedPenalty4 != -1 && test.expectedPenalty4 != penalty4 {
			ok = false
		}

		assert.True(t, ok)
	}
}
