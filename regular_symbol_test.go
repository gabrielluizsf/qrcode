package qrcode

import (
	"testing"

	"github.com/i9si-sistemas/assert"
	bitset "github.com/i9si-sistemas/bitset"
)

func TestBuildRegularSymbol(t *testing.T) {
	for k := range 7 {
		v := getQRCodeVersion(Low, 1)

		data := bitset.New()
		for range 26 {
			data.AppendNumBools(8, false)
		}

		_, err := buildRegularSymbol(*v, k, data, false)
		assert.NoError(t, err)
	}
}
