package encoding

import (
	"code.google.com/p/go.text/encoding/unicode"
	"golang.org/x/text/transform"
)

var UTF8 utf8Encoding = newUtf8Encoding("UTF8", unicode.IgnoreBOM)
var UTF8B utf8Encoding = newUtf8Encoding("UTF8B", unicode.ExpectBOM)

var ASCII utf8Encoding = newUtf8Encoding("ASCII", unicode.IgnoreBOM)

const (
	RuneError = '\uFFFD'     // the "error" Rune or "Unicode replacement character"
	RuneSelf  = 0x80         // characters below Runeself are represented as themselves in a single byte.
	MaxRune   = '\U0010FFFF' // Maximum valid Unicode code point.
	UTFMax    = 4            // maximum number of bytes of a UTF-8 encoded Unicode character.
)

// Code points in the surrogate range are not valid for UTF-8.
const (
	surrogateMin = 0xD800
	surrogateMax = 0xDFFF
)

const (
	t1 = 0x00 // 0000 0000
	tx = 0x80 // 1000 0000
	t2 = 0xC0 // 1100 0000
	t3 = 0xE0 // 1110 0000
	t4 = 0xF0 // 1111 0000
	t5 = 0xF8 // 1111 1000

	maskx = 0x3F // 0011 1111
	mask2 = 0x1F // 0001 1111
	mask3 = 0x0F // 0000 1111
	mask4 = 0x07 // 0000 0111

	rune1Max = 1<<7 - 1
	rune2Max = 1<<11 - 1
	rune3Max = 1<<16 - 1
)

type utf8Encoding struct {
	*splitter

	bom unicode.BOMPolicy

	decorder transform.Transformer

	name string
}

func (e utf8Encoding) String() string {
	return e.name
}

func newUtf8Encoding(name string, bom unicode.BOMPolicy) utf8Encoding {
	return utf8Encoding{
		splitter: &splitter{},
		bom:      bom,
		name:     name,
	}
}

func (e utf8Encoding) EncodingSearch(p []byte, atEOF bool) (nSrc int, err error, score int) {
	size := 0
	var r rune
	n := len(p)
	if n < 1 {
		err = transform.ErrShortSrc
		return 0, err, 0
	}
loop:
	for ; nSrc < n; nSrc += size {

		size = 1
		c0 := p[nSrc]

		if c0 == 0x00 {
			err = ErrInvalidEncoding
			break loop
		}

		// 1-byte, 7-bit sequence?
		if c0 < tx {
			//return rune(c0), 1, false
			continue
		}

		// unexpected continuation byte?
		if c0 < t2 {
			//return RuneError, 1, false
			err = ErrInvalidEncoding
			break loop
		}

		// need first continuation byte
		if n <= nSrc+1 {
			//return RuneError, 1, true
			err = transform.ErrShortSrc
			break loop
		}
		size = 2
		c1 := p[nSrc+1]
		if c1 < tx || t2 <= c1 {
			//return RuneError, 1, false
			err = ErrInvalidEncoding
			break loop
		}

		// 2-byte, 11-bit sequence?
		if c0 < t3 {
			r = rune(c0&mask2)<<6 | rune(c1&maskx)
			if r <= rune1Max {
				//return RuneError, 1, false
				err = ErrInvalidEncoding
				break loop
			}
			//return r, 2, false
			continue
		}

		// need second continuation byte
		if n <= nSrc+2 {
			//return RuneError, 1, true
			err = transform.ErrShortSrc
			break loop
		}
		size = 3
		c2 := p[nSrc+2]
		if c2 < tx || t2 <= c2 {
			//return RuneError, 1, false
			err = ErrInvalidEncoding
			break loop
		}

		// 3-byte, 16-bit sequence?
		if c0 < t4 {
			r = rune(c0&mask3)<<12 | rune(c1&maskx)<<6 | rune(c2&maskx)
			if r <= rune2Max {
				//return RuneError, 1, false
				err = ErrInvalidEncoding
				break loop
			}
			if surrogateMin <= r && r <= surrogateMax {
				//return RuneError, 1, false
				err = ErrInvalidEncoding
				break loop
			}
			//return r, 3, false
			continue
		}

		// need third continuation byte
		if n <= nSrc+3 {
			//return RuneError, 1, true
			err = transform.ErrShortSrc
			break loop
		}
		size = 4
		c3 := p[nSrc+3]
		if c3 < tx || t2 <= c3 {
			//return RuneError, 1, false
			err = ErrInvalidEncoding
			break loop
		}

		// 4-byte, 21-bit sequence?
		if c0 < t5 {
			r = rune(c0&mask4)<<18 | rune(c1&maskx)<<12 | rune(c2&maskx)<<6 | rune(c3&maskx)
			if r <= rune3Max || MaxRune < r {
				//return RuneError, 1, false
				err = ErrInvalidEncoding
				break loop
			}
			//return r, 4, false
			continue
		}

		// error
		//return RuneError, 1, false
		err = ErrInvalidEncoding
		break loop
	}

	if atEOF && err == transform.ErrShortSrc {
		err = ErrInvalidEncoding
	}
	return nSrc, err, score
}

func (c utf8Encoding) Decode(b []byte) (string, error) {
	return string(b), nil
}

func (c utf8Encoding) Encode(s string) ([]byte, error) {
	return []byte(s), nil
}