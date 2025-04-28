package qrcode

import (
	"testing"

	"github.com/i9si-sistemas/assert"
	bitset "github.com/i9si-sistemas/bitset"
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

func TestSingleModeEncodings(t *testing.T) {
	tests := []struct {
		dataEncoderType dataEncoderType
		dataMode        dataMode
		data            string
		expected        *bitset.Bitset
	}{
		{
			dataEncoderType1To9,
			dataModeNumeric,
			"01234567",
			bitset.NewFromBase2String("0001 0000001000 0000001100 0101011001 1000011"),
		},
		{
			dataEncoderType1To9,
			dataModeAlphanumeric,
			"AC-42",
			bitset.NewFromBase2String("0010 000000101 00111001110 11100111001 000010"),
		},
		{
			dataEncoderType1To9,
			dataModeByte,
			"123",
			bitset.NewFromBase2String("0100 00000011 00110001 00110010 00110011"),
		},
		{
			dataEncoderType10To26,
			dataModeByte,
			"123",
			bitset.NewFromBase2String("0100 00000000 00000011 00110001 00110010 00110011"),
		},
		{
			dataEncoderType27To40,
			dataModeByte,
			"123",
			bitset.NewFromBase2String("0100 00000000 00000011 00110001 00110010 00110011"),
		},
	}

	for _, test := range tests {
		encoder := newDataEncoder(test.dataEncoderType)
		encoded := bitset.New()

		encoder.encodeDataRaw([]byte(test.data), test.dataMode, encoded)

		assert.True(t, test.expected.Equals(encoded))
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
