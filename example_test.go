package qrcode

import (
	"image/color"
	"os"
	"testing"

	"github.com/i9si-sistemas/assert"
)

const i9siDomain = "https://www.i9sisistemas.com.br/"

func TestExampleEncode(t *testing.T) {
	_, err := Encode(i9siDomain, Medium, 256)
	assert.NoError(t, err)
}

func TestExampleWriteFile(t *testing.T) {
	filename := "example.png"
	if err := WriteFile(i9siDomain, Medium, 256, filename); err != nil {
		err = os.Remove(filename)
		assert.NoError(t, err)
	}
}

func TestExampleEncodeWithColourAndWithoutBorder(t *testing.T) {
	q, err := New(i9siDomain, Highest)
	assert.NoError(t, err)

	foregroundColor := color.RGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xff}
	backgroundColor := color.RGBA{R: 0xef, G: 0xef, B: 0xef, A: 0xff}

	err = q.WithNoBorder().WithColors(
		foregroundColor, 
		backgroundColor,
	).WriteFile(256, "example2.png")
	assert.NoError(t, err)
	err = q.WriteFileWithoutSize("example3.png")
	assert.NoError(t, err)
}

func TestExampleWriteFileWithoutSize(t *testing.T) {
	filename := "example4.png"
	if err := WriteFile(i9siDomain, Medium, 0, filename); err != nil {
		err = os.Remove(filename)
		assert.NoError(t, err)
	}
}
