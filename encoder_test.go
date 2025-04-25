package qrcode

import (
	"testing"

	"github.com/i9si-sistemas/assert"
)

func TestClassifyDataMode(t *testing.T) {
	tests := []struct {
		data   []byte
		actual []segment
	}{
		{
			[]byte{0x30},
			[]segment{
				{
					dataModeNumeric,
					[]byte{0x30},
				},
			},
		},
		{
			[]byte{0x30, 0x41, 0x42, 0x43, 0x20, 0x00, 0xf0, 0xf1, 0xf2, 0x31},
			[]segment{
				{
					dataModeNumeric,
					[]byte{0x30},
				},
				{
					dataModeAlphanumeric,
					[]byte{0x41, 0x42, 0x43, 0x20},
				},
				{
					dataModeByte,
					[]byte{0x00, 0xf0, 0xf1, 0xf2},
				},
				{
					dataModeNumeric,
					[]byte{0x31},
				},
			},
		},
	}

	for _, test := range tests {
		encoder := newDataEncoder(dataEncoderType1To9)
		encoder.encode(test.data)

		assert.Equal(t, test.actual, encoder.actual)
	}
}

func TestByteModeLengthCalculations(t *testing.T) {
	tests := []struct {
		dataEncoderType dataEncoderType
		dataMode        dataMode
		numSymbols      int
		expectedLength  int
	}{}

	for _, test := range tests {
		encoder := newDataEncoder(test.dataEncoderType)
		var resultLength int

		resultLength, err := encoder.encodedLength(test.dataMode, test.numSymbols)
		if test.expectedLength == -1 {
			assert.NotNil(t, err)
		}
		assert.Equal(t, resultLength, test.expectedLength)
	}
}
