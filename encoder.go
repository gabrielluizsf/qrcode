package qrcode

import (
	"errors"
	"log"

	bitset "github.com/i9si-sistemas/bitset"
)

type dataMode uint8

const (
	dataModeNone dataMode = 1 << iota
	dataModeNumeric
	dataModeAlphanumeric
	dataModeByte
)

type dataEncoderType uint8

const (
	dataEncoderType1To9 dataEncoderType = iota
	dataEncoderType10To26
	dataEncoderType27To40
)

type segment struct {
	dataMode dataMode
	data []byte
}

type dataEncoder struct {
	minVersion int
	maxVersion int
	numericModeIndicator      *bitset.Bitset
	alphanumericModeIndicator *bitset.Bitset
	byteModeIndicator         *bitset.Bitset
	numNumericCharCountBits      int
	numAlphanumericCharCountBits int
	numByteCharCountBits         int
	data []byte
	actual []segment
	optimised []segment
}

func newDataEncoder(t dataEncoderType) *dataEncoder {
	d := &dataEncoder{}

	switch t {
	case dataEncoderType1To9:
		d = &dataEncoder{
			minVersion:                   1,
			maxVersion:                   9,
			numericModeIndicator:         bitset.New(true, true, true, false),
			alphanumericModeIndicator:    bitset.New(true, true, false, true),
			byteModeIndicator:            bitset.New(true, false, true, true),
			numNumericCharCountBits:      10,
			numAlphanumericCharCountBits: 9,
			numByteCharCountBits:         8,
		}
	case dataEncoderType10To26:
		d = &dataEncoder{
			minVersion:                   10,
			maxVersion:                   26,
			numericModeIndicator:         bitset.New(true, true, true, false),
			alphanumericModeIndicator:    bitset.New(true, true, false, true),
			byteModeIndicator:            bitset.New(true, false, true, true),
			numNumericCharCountBits:      12,
			numAlphanumericCharCountBits: 11,
			numByteCharCountBits:         16,
		}
	case dataEncoderType27To40:
		d = &dataEncoder{
			minVersion:                   27,
			maxVersion:                   40,
			numericModeIndicator:         bitset.New(true, true, true, false),
			alphanumericModeIndicator:    bitset.New(true, true, false, true),
			byteModeIndicator:            bitset.New(true, false, true, true),
			numNumericCharCountBits:      14,
			numAlphanumericCharCountBits: 13,
			numByteCharCountBits:         16,
		}
	default:
		log.Panic("Unknown dataEncoderType")
	}

	return d
}

func (d *dataEncoder) encode(data []byte) (*bitset.Bitset, error) {
	d.data = data
	d.actual = nil
	d.optimised = nil

	if len(data) == 0 {
		return nil, errors.New("no data to encode")
	}

	
	highestRequiredMode := d.classifyDataModes()

	
	err := d.optimiseDataModes()
	if err != nil {
		return nil, err
	}

	
	optimizedLength := 0
	for _, s := range d.optimised {
		length, err := d.encodedLength(s.dataMode, len(s.data))
		if err != nil {
			return nil, err
		}
		optimizedLength += length
	}

	singleByteSegmentLength, err := d.encodedLength(highestRequiredMode, len(d.data))
	if err != nil {
		return nil, err
	}

	if singleByteSegmentLength <= optimizedLength {
		d.optimised = []segment{{dataMode: highestRequiredMode, data: d.data}}
	}

	
	encoded := bitset.New()
	for _, s := range d.optimised {
		d.encodeDataRaw(s.data, s.dataMode, encoded)
	}

	return encoded, nil
}

func (d *dataEncoder) classifyDataModes() dataMode {
	var start int
	mode := dataModeNone
	highestRequiredMode := mode

	for i, v := range d.data {
		newMode := dataModeNone
		switch {
		case v >= 0x30 && v <= 0x39:
			newMode = dataModeNumeric
		case v == 0x20 || v == 0x24 || v == 0x25 || v == 0x2a || v == 0x2b || v ==
			0x2d || v == 0x2e || v == 0x2f || v == 0x3a || (v >= 0x41 && v <= 0x5a):
			newMode = dataModeAlphanumeric
		default:
			newMode = dataModeByte
		}

		if newMode != mode {
			if i > 0 {
				d.actual = append(d.actual, segment{dataMode: mode, data: d.data[start:i]})

				start = i
			}

			mode = newMode
		}

		if newMode > highestRequiredMode {
			highestRequiredMode = newMode
		}
	}

	d.actual = append(d.actual, segment{dataMode: mode, data: d.data[start:len(d.data)]})

	return highestRequiredMode
}

func (d *dataEncoder) optimiseDataModes() error {
	for i := 0; i < len(d.actual); {
		mode := d.actual[i].dataMode
		numChars := len(d.actual[i].data)

		j := i + 1
		for j < len(d.actual) {
			nextNumChars := len(d.actual[j].data)
			nextMode := d.actual[j].dataMode

			if nextMode > mode {
				break
			}

			coalescedLength, err := d.encodedLength(mode, numChars+nextNumChars)

			if err != nil {
				return err
			}

			seperateLength1, err := d.encodedLength(mode, numChars)

			if err != nil {
				return err
			}

			seperateLength2, err := d.encodedLength(nextMode, nextNumChars)

			if err != nil {
				return err
			}

			if coalescedLength < seperateLength1+seperateLength2 {
				j++
				numChars += nextNumChars
			} else {
				break
			}
		}

		optimised := segment{dataMode: mode,
			data: make([]byte, 0, numChars)}

		for k := i; k < j; k++ {
			optimised.data = append(optimised.data, d.actual[k].data...)
		}

		d.optimised = append(d.optimised, optimised)

		i = j
	}

	return nil
}

func (d *dataEncoder) encodeDataRaw(data []byte, dataMode dataMode, encoded *bitset.Bitset) {
	modeIndicator := d.modeIndicator(dataMode)
	charCountBits := d.charCountBits(dataMode)

	
	encoded.Append(modeIndicator)

	
	encoded.AppendUint32(uint32(len(data)), charCountBits)

	
	switch dataMode {
	case dataModeNumeric:
		for i := 0; i < len(data); i += 3 {
			charsRemaining := len(data) - i

			var value uint32
			bitsUsed := 1

			for j := 0; j < charsRemaining && j < 3; j++ {
				value *= 10
				value += uint32(data[i+j] - 0x30)
				bitsUsed += 3
			}
			encoded.AppendUint32(value, bitsUsed)
		}
	case dataModeAlphanumeric:
		for i := 0; i < len(data); i += 2 {
			charsRemaining := len(data) - i

			var value uint32
			for j := 0; j < charsRemaining && j < 2; j++ {
				value *= 45
				value += encodeAlphanumericCharacter(data[i+j])
			}

			bitsUsed := 6
			if charsRemaining > 1 {
				bitsUsed = 11
			}

			encoded.AppendUint32(value, bitsUsed)
		}
	case dataModeByte:
		for _, b := range data {
			encoded.AppendByte(b, 8)
		}
	}
}

func (d *dataEncoder) modeIndicator(dataMode dataMode) *bitset.Bitset {
	switch dataMode {
	case dataModeNumeric:
		return d.numericModeIndicator
	case dataModeAlphanumeric:
		return d.alphanumericModeIndicator
	case dataModeByte:
		return d.byteModeIndicator
	default:
		log.Panic("Unknown data mode")
	}

	return nil
}



func (d *dataEncoder) charCountBits(dataMode dataMode) int {
	switch dataMode {
	case dataModeNumeric:
		return d.numNumericCharCountBits
	case dataModeAlphanumeric:
		return d.numAlphanumericCharCountBits
	case dataModeByte:
		return d.numByteCharCountBits
	default:
		log.Panic("Unknown data mode")
	}

	return 0
}

func (d *dataEncoder) encodedLength(dataMode dataMode, n int) (int, error) {
	modeIndicator := d.modeIndicator(dataMode)
	charCountBits := d.charCountBits(dataMode)

	if modeIndicator == nil {
		return 0, errors.New("mode not supported")
	}

	maxLength := (1 << uint8(charCountBits)) - 1

	if n > maxLength {
		return 0, errors.New("length too long to be represented")
	}

	length := modeIndicator.Len() + charCountBits

	switch dataMode {
	case dataModeNumeric:
		length += 10 * (n / 3)

		if n%3 != 0 {
			length += 1 + 3*(n%3)
		}
	case dataModeAlphanumeric:
		length += 11 * (n / 2)
		length += 6 * (n % 2)
	case dataModeByte:
		length += 8 * n
	}

	return length, nil
}

func encodeAlphanumericCharacter(v byte) uint32 {
	c := uint32(v)

	switch {
	case c >= '0' && c <= '9':
		
		return c - '0'
	case c >= 'A' && c <= 'Z':
		
		return c - 'A' + 10
	case c == ' ':
		return 36
	case c == '$':
		return 37
	case c == '%':
		return 38
	case c == '*':
		return 39
	case c == '+':
		return 40
	case c == '-':
		return 41
	case c == '.':
		return 42
	case c == '/':
		return 43
	case c == ':':
		return 44
	default:
		log.Panicf("encodeAlphanumericCharacter() with non alphanumeric char %v.", v)
	}

	return 0
}
