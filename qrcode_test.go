package qrcode

import (
	"strings"
	"testing"
)

func TestQRCodeMaxCapacity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TestQRCodeCapacity")
	}

	tests := []struct {
		string         string
		numRepetitions int
	}{
		{
			"0",
			7089,
		},
		{
			"A",
			4296,
		},
		{
			"#",
			2953,
		},
		{
			"#1",
			1476,
		},
	}

	for _, test := range tests {
		_, err := New(strings.Repeat(test.string, test.numRepetitions), Low)

		if err != nil {
			t.Errorf("%d x '%s' got %s expected success", test.numRepetitions,
				test.string, err.Error())
		}
	}

	for _, test := range tests {
		_, err := New(strings.Repeat(test.string, test.numRepetitions+1), Low)

		if err == nil {
			t.Errorf("%d x '%s' chars encodable, expected not encodable",
				test.numRepetitions+1, test.string)
		}
	}
}

func TestQRCodeVersionCapacity(t *testing.T) {
	tests := []struct {
		version         int
		level           RecoveryLevel
		maxNumeric      int
		maxAlphanumeric int
		maxByte         int
	}{
		{
			1,
			Low,
			41,
			25,
			17,
		},
		{
			2,
			Low,
			77,
			47,
			32,
		},
		{
			2,
			Highest,
			34,
			20,
			14,
		},
		{
			40,
			Low,
			7089,
			4296,
			2953,
		},
		{
			40,
			Highest,
			3057,
			1852,
			1273,
		},
	}

	for i, test := range tests {
		numericData := strings.Repeat("1", test.maxNumeric)
		alphanumericData := strings.Repeat("A", test.maxAlphanumeric)
		byteData := strings.Repeat("#", test.maxByte)

		var n *QRCode
		var a *QRCode
		var b *QRCode
		var err error

		n, err = New(numericData, test.level)
		if err != nil {
			t.Fatal(err.Error())
		}

		a, err = New(alphanumericData, test.level)
		if err != nil {
			t.Fatal(err.Error())
		}

		b, err = New(byteData, test.level)
		if err != nil {
			t.Fatal(err.Error())
		}

		if n.VersionNumber != test.version {
			t.Fatalf("Test #%d numeric has version #%d, expected #%d", i,
				n.VersionNumber, test.version)
		}

		if a.VersionNumber != test.version {
			t.Fatalf("Test #%d alphanumeric has version #%d, expected #%d", i,
				a.VersionNumber, test.version)
		}

		if b.VersionNumber != test.version {
			t.Fatalf("Test #%d byte has version #%d, expected #%d", i,
				b.VersionNumber, test.version)
		}
	}
}

func BenchmarkQRCodeURLSize(b *testing.B) {
	for b.Loop() {
		New("http://www.example.org", Medium)
	}
}

func BenchmarkQRCodeMaximumSize(b *testing.B) {
	for b.Loop() {
		New(strings.Repeat("0", 7089), Low)
	}
}
