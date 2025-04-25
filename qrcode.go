package qrcode

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/i9si-sistemas/bitset"
	"github.com/i9si-sistemas/reedsolomon"
)

type QRCode struct {
	Content         string
	Level           RecoveryLevel
	VersionNumber   int
	ForegroundColor color.Color
	BackgroundColor color.Color
	DisableBorder   bool
	encoder         *dataEncoder
	version         qrCodeVersion
	data            *bitset.Bitset
	symbol          *symbol
	mask            int
}

// New returns a new QRCode.
func New(content string, level RecoveryLevel) (*QRCode, error) {
	encoders := []dataEncoderType{dataEncoderType1To9, dataEncoderType10To26,
		dataEncoderType27To40}

	var (
		encoder       *dataEncoder
		encoded       *bitset.Bitset
		chosenVersion *qrCodeVersion
		err           error
	)

	for _, t := range encoders {
		encoder = newDataEncoder(t)
		encoded, err = encoder.encode([]byte(content))
		if err != nil {
			continue
		}
		chosenVersion = chooseQRCodeVersion(level, encoder, encoded.Len())

		if chosenVersion != nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	if chosenVersion == nil {
		return nil, errors.New("content too long to encode")
	}

	q := &QRCode{
		Content: content,

		Level:         level,
		VersionNumber: chosenVersion.version,

		ForegroundColor: color.Black,
		BackgroundColor: color.White,

		encoder: encoder,
		data:    encoded,
		version: *chosenVersion,
	}

	return q, nil
}

// NewWithForcedVersion returns a new QRCode with a forced version.
func NewWithForcedVersion(content string, version int, level RecoveryLevel) (*QRCode, error) {
	var encoder *dataEncoder

	switch {
	case version >= 1 && version <= 9:
		encoder = newDataEncoder(dataEncoderType1To9)
	case version >= 10 && version <= 26:
		encoder = newDataEncoder(dataEncoderType10To26)
	case version >= 27 && version <= 40:
		encoder = newDataEncoder(dataEncoderType27To40)
	default:
		return nil, fmt.Errorf("invalid version %d (expected 1-40 inclusive)", version)
	}

	var encoded *bitset.Bitset
	encoded, err := encoder.encode([]byte(content))

	if err != nil {
		return nil, err
	}

	chosenVersion := getQRCodeVersion(level, version)

	if chosenVersion == nil {
		return nil, errors.New("cannot find QR Code version")
	}

	if encoded.Len() > chosenVersion.numDataBits() {
		return nil, fmt.Errorf("cannot encode QR code: content too large for fixed size QR Code version %d (encoded length is %d bits, maximum length is %d bits)",
			version,
			encoded.Len(),
			chosenVersion.numDataBits())
	}

	q := &QRCode{
		Content:         content,
		Level:           level,
		VersionNumber:   chosenVersion.version,
		ForegroundColor: color.Black,
		BackgroundColor: color.White,
		encoder:         encoder,
		data:            encoded,
		version:         *chosenVersion,
	}

	return q, nil
}

// Encode returns a PNG image of the QRCode.
func Encode(content string, level RecoveryLevel, size int) ([]byte, error) {
	q, err := New(content, level)
	if err != nil {
		return nil, err
	}

	b, err := q.PNG(size)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Write writes a PNG image of the QRCode
func WriteFile(content string, level RecoveryLevel, size int, filename string) error {
	q, err := New(content, level)
	if err != nil {
		return err
	}

	return q.WriteFile(size, filename)
}

// WriteColorFile writes a PNG image of the QRCode with custom foreground and background colors.
func WriteColorFile(
	content string,
	level RecoveryLevel,
	size int,
	background, foreground color.Color,
	filename string,
) error {
	q, err := New(content, level)
	if err != nil {
		return err
	}
	q.BackgroundColor = background
	q.ForegroundColor = foreground

	return q.WriteFile(size, filename)
}

// Bitmap returns a two-dimensional array of booleans representing the QRCode.
func (q *QRCode) Bitmap() [][]bool {
	q.encode()

	return q.symbol.bitmap()
}

// Image returns an image.Image of the QRCode.
func (q *QRCode) Image(size int) image.Image {
	q.encode()

	realSize := q.symbol.size

	if size < 0 {
		size = size * -1 * realSize
	}

	if size < realSize {
		size = realSize
	}

	rect := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{size, size}}

	p := color.Palette([]color.Color{q.BackgroundColor, q.ForegroundColor})
	img := image.NewPaletted(rect, p)
	fgClr := uint8(img.Palette.Index(q.ForegroundColor))

	bitmap := q.symbol.bitmap()

	modulesPerPixel := float64(realSize) / float64(size)
	for y := range size {
		y2 := int(float64(y) * modulesPerPixel)
		for x := range size {
			x2 := int(float64(x) * modulesPerPixel)
			v := bitmap[y2][x2]
			if v {
				pos := img.PixOffset(x, y)
				img.Pix[pos] = fgClr
			}
		}
	}

	return img
}

// PNG returns a PNG image of the QRCode.
func (q *QRCode) PNG(size int) ([]byte, error) {
	img := q.Image(size)	
	buf := new(bytes.Buffer)
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	err := encoder.Encode(buf, img)
	if err != nil {
		return q.PNG(len(buf.Bytes()))
	}

	return buf.Bytes(), nil
}

// Write writes a PNG image of the QRCode.
func (q *QRCode) Write(size int, out io.Writer) error {
	png, err := q.PNG(size)
	if err != nil {
		return err
	}
	if _, err = out.Write(png); err != nil {
		return err
	}
	return err
}

// WriteFile writes a PNG image of the QRCode.
func (q *QRCode) WriteFile(size int, filename string) error {
	png, err := q.PNG(size)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, png, os.FileMode(0644))
}

func (q *QRCode) encode() {
	numTerminatorBits := q.version.numTerminatorBitsRequired(q.data.Len())

	q.addTerminatorBits(numTerminatorBits)
	q.addPadding()

	encoded := q.encodeBlocks()

	const numMasks int = 8
	penalty := 0

	for mask := range numMasks {
		s, err := buildRegularSymbol(q.version, mask, encoded, !q.DisableBorder)
		if err != nil {
			log.Panic(err.Error())
		}

		numEmptyModules := s.numEmptyModules()
		if numEmptyModules != 0 {
			log.Panicf("bug: numEmptyModules is %d (expected 0) (version=%d)",
				numEmptyModules, q.VersionNumber)
		}

		p := s.penaltyScore()

		if q.symbol == nil || p < penalty {
			q.symbol = s
			q.mask = mask
			penalty = p
		}
	}
}

func (q *QRCode) addTerminatorBits(numTerminatorBits int) {
	q.data.AppendNumBools(numTerminatorBits, false)
}

func (q *QRCode) encodeBlocks() *bitset.Bitset {
	type dataBlock struct {
		data          *bitset.Bitset
		ecStartOffset int
	}

	block := make([]dataBlock, q.version.numBlocks())

	start := 0
	end := 0
	blockID := 0

	for _, b := range q.version.block {
		for range b.numBlocks {
			start = end
			end = start + b.numDataCodewords*8

			numErrorCodewords := b.numCodewords - b.numDataCodewords
			block[blockID].data = reedsolomon.Encode(q.data.Substr(start, end), numErrorCodewords)
			block[blockID].ecStartOffset = end - start

			blockID++
		}
	}

	result := bitset.New()

	working := true
	for i := 0; working; i += 8 {
		working = false

		for j, b := range block {
			if i >= block[j].ecStartOffset {
				continue
			}

			result.Append(b.data.Substr(i, i+8))

			working = true
		}
	}

	working = true
	for i := 0; working; i += 8 {
		working = false

		for j, b := range block {
			offset := i + block[j].ecStartOffset
			if offset >= block[j].data.Len() {
				continue
			}

			result.Append(b.data.Substr(offset, offset+8))

			working = true
		}
	}

	result.AppendNumBools(q.version.numRemainderBits, false)

	return result
}

func (q *QRCode) addPadding() {
	numDataBits := q.version.numDataBits()

	if q.data.Len() == numDataBits {
		return
	}

	q.data.AppendNumBools(q.version.numBitsToPadToCodeword(q.data.Len()), false)

	padding := [2]*bitset.Bitset{
		bitset.New(true, true, true, false, true, true, false, false),
		bitset.New(false, false, false, true, false, false, false, true),
	}

	i := 0
	for numDataBits-q.data.Len() >= 8 {
		q.data.Append(padding[i])

		i = 1 - i
	}

	if q.data.Len() != numDataBits {
		log.Panicf("BUG: got len %d, expected %d", q.data.Len(), numDataBits)
	}
}

// ToString returns a string representation of the QRCode.
func (q *QRCode) ToString(inverseColor bool) string {
	bits := q.Bitmap()
	var buf bytes.Buffer
	for y := range bits {
		for x := range bits[y] {
			if bits[y][x] != inverseColor {
				buf.WriteString("  ")
			} else {
				buf.WriteString("██")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// ToSmallString returns a small string representation of the QRCode.
func (q *QRCode) ToSmallString(inverseColor bool) string {
	bits := q.Bitmap()
	var buf bytes.Buffer
	for y := 0; y < len(bits)-1; y += 2 {
		for x := range bits[y] {
			if bits[y][x] == bits[y+1][x] {
				if bits[y][x] != inverseColor {
					buf.WriteString(" ")
				} else {
					buf.WriteString("█")
				}
			} else {
				if bits[y][x] != inverseColor {
					buf.WriteString("▄")
				} else {
					buf.WriteString("▀")
				}
			}
		}
		buf.WriteString("\n")
	}
	if len(bits)%2 == 1 {
		y := len(bits) - 1
		for x := range bits[y] {
			if bits[y][x] != inverseColor {
				buf.WriteString(" ")
			} else {
				buf.WriteString("▀")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}
