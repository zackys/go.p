package encoding

import (
	"bytes"
	"code.google.com/p/go.text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"unicode/utf8"
)

var ShiftJIS *shiftJIS = newShiftJIS()

type shiftJIS struct {
	shiftJISDecoder
	*splitter

	decoder transform.Transformer
}

func (shiftJIS) String() string {
	return "Shift JIS"
}

func (shiftJIS) NewEncodingSearcher() EncodingSearcher {
	return shiftJISDecoder{}
}

func newShiftJIS() *shiftJIS {
	return &shiftJIS{
		shiftJISDecoder: shiftJISDecoder{},
		splitter:        &splitter{},
	}
}

type shiftJISDecoder struct{ transform.NopResetter }

func (shiftJISDecoder) EncodingSearch(src []byte, atEOF bool) (nSrc int, err error, score int) {
	//r, size := rune(0), 0
	size := 0
loop:
	for ; nSrc < len(src); nSrc += size {
		switch c0 := src[nSrc]; {
		case c0 == 0x00:
			err = ErrInvalidEncoding
			break loop
		case c0 < utf8.RuneSelf:
			//r, size = rune(c0), 1
			size = 1

		case 0xa1 <= c0 && c0 < 0xe0:
			//r, size = rune(c0)+(0xff61-0xa1), 1
			size = 1

		case (0x81 <= c0 && c0 < 0xa0) || (0xe0 <= c0 && c0 < 0xf0):
			if c0 <= 0x9f {
				c0 -= 0x70
			} else {
				c0 -= 0xb0
			}
			c0 = 2*c0 - 0x21

			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			switch {
			case c1 < 0x40:
				err = ErrInvalidEncoding
				break loop
			case c1 < 0x7f:
				c0--
				c1 -= 0x40
			case c1 == 0x7f:
				err = ErrInvalidEncoding
				break loop
			case c1 < 0x9f:
				c0--
				c1 -= 0x41
			case c1 < 0xfd:
				c1 -= 0x9f
			default:
				err = ErrInvalidEncoding
				break loop
			}
			//r, size = '\ufffd', 2
			score++
			size = 2
			//			if i := int(c0)*94 + int(c1); i < len(jis0208Decode) {
			//				r = rune(jis0208Decode[i])
			//				if r == 0 {
			//					r = '\ufffd'
			//				}
			//			}

		default:
			err = ErrInvalidEncoding
			break loop
		}

		//		if nDst+utf8.RuneLen(r) > len(dst) {
		//			err = transform.ErrShortDst
		//			break loop
		//		}
		//		nDst += utf8.EncodeRune(dst[nDst:], r)
	}
	if atEOF && err == transform.ErrShortSrc {
		err = ErrInvalidEncoding
	}
	return nSrc, err, score
}

func (c *shiftJIS) getDecoder() transform.Transformer {
	if c.decoder == nil {
		c.decoder = japanese.ShiftJIS.NewDecoder()
	} else {
		c.decoder.Reset()
	}

	return c.decoder
}

func (c *shiftJIS) Decode(b []byte) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(b), c.getDecoder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}
