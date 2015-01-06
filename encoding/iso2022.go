package encoding

import (
	"bytes"
	"code.google.com/p/go.text/encoding/japanese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"strings"
	"unicode/utf8"
)

var ISO2022JP *iso2022JPEncoding = newIso2022JPEncoding()

const (
	asciiState = iota
	katakanaState
	jis0208State
	jis0212State
)

const asciiEsc = 0x1b

type iso2022JPDecorder int

type iso2022JPEncoding struct {
	*iso2022JPDecorder
	*splitter

	decoder transform.Transformer
	encoder transform.Transformer
}

func (iso2022JPEncoding) String() string {
	return "ISO2022"
}

func newIso2022JPEncoding() *iso2022JPEncoding {
	return &iso2022JPEncoding{
		iso2022JPDecorder: new(iso2022JPDecorder),
		splitter:          &splitter{},
	}
}

func (d *iso2022JPDecorder) Reset() {
	*d = asciiState
}

func (d *iso2022JPDecorder) EncodingSearch(src []byte, atEOF bool) (nSrc int, err error, score int) {
	size := 0
loop:
	for ; nSrc < len(src); nSrc += size {
		c0 := src[nSrc]

		if c0 == 0x00 {
			err = ErrInvalidEncoding
			break loop
		}

		if c0 >= utf8.RuneSelf {
			err = ErrInvalidEncoding
			break loop
		}

		if c0 == asciiEsc {
			if nSrc+2 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			size = 3
			c1 := src[nSrc+1]
			c2 := src[nSrc+2]
			switch {
			case c1 == '$' && (c2 == '@' || c2 == 'B'):
				*d = jis0208State
				score++
				continue
			case c1 == '$' && c2 == '(':
				if nSrc+3 >= len(src) {
					err = transform.ErrShortSrc
					break loop
				}
				size = 4
				if src[nSrc+3] == 'D' {
					*d = jis0212State
					score++
					continue
				}
			case c1 == '(' && (c2 == 'B' || c2 == 'J'):
				*d = asciiState
				score++
				continue
			case c1 == '(' && c2 == 'I':
				*d = katakanaState
				score++
				continue
			}
			err = ErrInvalidEncoding
			break loop
		}

		switch *d {
		case asciiState:
			//r, size = rune(c0), 1
			size = 1

		case katakanaState:
			if c0 < 0x21 || 0x60 <= c0 {
				err = ErrInvalidEncoding
				break loop
			}
			//r, size = rune(c0)+(0xff61-0x21), 1
			size = 1

		default:
			if c0 == 0x0a {
				*d = asciiState
				//r, size = rune(c0), 1
				size = 1
				break
			}
			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			size = 2
		}

	}
	if atEOF && err == transform.ErrShortSrc {
		err = ErrInvalidEncoding
	}
	return nSrc, err, score
}

func (c *iso2022JPEncoding) getDecoder() transform.Transformer {
	if c.decoder == nil {
		c.decoder = japanese.ISO2022JP.NewDecoder()
	} else {
		c.decoder.Reset()
	}

	return c.decoder
}

func (c *iso2022JPEncoding) Decode(b []byte) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(b), c.getDecoder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}

func (c *iso2022JPEncoding) getEncoder() transform.Transformer {
	if c.encoder == nil {
		c.encoder = japanese.ISO2022JP.NewEncoder()
	} else {
		c.encoder.Reset()
	}

	return c.encoder
}

func (c *iso2022JPEncoding) Encode(s string) ([]byte, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(s), c.getEncoder()))
	if err != nil {
		return nil, err
	}
	return ret, err
}
