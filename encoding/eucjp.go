package encoding

import (
	"bytes"
	"code.google.com/p/go.text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"unicode/utf8"
)

// EUCJP is the EUC-JP encoding.
var EUCJP eucJP = newEucJP()

type eucJP struct {
	eucJPDecoder

	*splitter
	decorder transform.Transformer
}

func (eucJP) String() string {
	return "EUC-JP"
}

func newEucJP() eucJP {
	return eucJP{
		eucJPDecoder: eucJPDecoder{},
		splitter:     &splitter{},
	}
}

type eucJPDecoder struct{ transform.NopResetter }

func (eucJPDecoder) EncodingSearch(src []byte, atEOF bool) (nSrc int, err error, score int) {
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

		case c0 == 0x8e:
			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			if c1 < 0xa1 || 0xdf < c1 {
				err = ErrInvalidEncoding
				break loop
			}
			//r, size = rune(c1)+(0xff61-0xa1), 2
			score++
			size = 2

		case c0 == 0x8f:
			if nSrc+2 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			if c1 < 0xa1 || 0xfe < c1 {
				err = ErrInvalidEncoding
				break loop
			}
			c2 := src[nSrc+2]
			if c2 < 0xa1 || 0xfe < c2 {
				err = ErrInvalidEncoding
				break loop
			}
			//r, size = '\ufffd', 3
			score++
			size = 3
			//			if i := int(c1-0xa1)*94 + int(c2-0xa1); i < len(jis0212Decode) {
			//				r = rune(jis0212Decode[i])
			//				if r == 0 {
			//					r = '\ufffd'
			//				}
			//			}

		case 0xa1 <= c0 && c0 <= 0xfe:
			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			if c1 < 0xa1 || 0xfe < c1 {
				err = ErrInvalidEncoding
				break loop
			}
			//r, size = '\ufffd', 2
			score++
			size = 2
			//			if i := int(c0-0xa1)*94 + int(c1-0xa1); i < len(jis0208Decode) {
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

func (c eucJP) getDecorder() transform.Transformer {
	if c.decorder == nil {
		c.decorder = japanese.EUCJP.NewDecoder()
	}

	return c.decorder
}

func (c eucJP) Decode(b []byte) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(b), c.getDecorder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}
