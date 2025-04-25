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
