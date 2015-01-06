package encoding

import (
	"bytes"
	"code.google.com/p/go.text/encoding/unicode"
	"container/list"
	"golang.org/x/text/transform"
	"io/ioutil"
	"strings"
)

var UTF16   *utf16Encoding = newUtf16Encoding("UTF16",   unicode.LittleEndian, unicode.ExpectBOM)
var UTF16B  *utf16Encoding = newUtf16Encoding("UTF16B",  unicode.BigEndian, unicode.ExpectBOM)
var UTF16LE *utf16Encoding = newUtf16Encoding("UTF16LE", unicode.LittleEndian, unicode.IgnoreBOM)
var UTF16BE *utf16Encoding = newUtf16Encoding("UTF16BE", unicode.BigEndian, unicode.IgnoreBOM)

const (
	// 0xd800-0xdc00 encodes the high 10 bits of a pair.
	// 0xdc00-0xe000 encodes the low 10 bits of a pair.
	// the value is those 20 bits plus 0x10000.
	surr1 = 0xd8
	surr2 = 0xdc
	surr3 = 0xe0

	surrSelf = 0x10000
)

type utf16Encoding struct {
	*utf16Decoder
	*utf16splitter

	endian unicode.Endianness
	bom    unicode.BOMPolicy

	decoder transform.Transformer
	encoder transform.Transformer

	name string
}

func (e utf16Encoding) String() string {
	return e.name
}

func newUtf16Encoding(name string, endian unicode.Endianness, bom unicode.BOMPolicy) *utf16Encoding {
	return &utf16Encoding{
		utf16Decoder:  &utf16Decoder{},
		utf16splitter: &utf16splitter{endian: endian},
		endian:        endian,
		bom:           bom,
		name:          name,
	}
}

type utf16Decoder struct {
	scoreLE int
	scoreBE int
}

func (d *utf16Decoder) EncodingSearch(p []byte, atEOF bool) (nSrc int, err error, score int) {
	size := 2
	n := len(p)
	if n < 1 {
		err = transform.ErrShortSrc
		return 0, err, 0
	}
loop:
	for ; nSrc < n; nSrc += size {

		c0 := p[nSrc]

		// need first continuation byte
		if n <= nSrc+1 {
			//return RuneError, 1, true
			err = transform.ErrShortSrc
			break loop
		}
		c1 := p[nSrc+1]

		if c0 == 0x00 && (0x03 < c1 && c1 < 0x20) {
			d.scoreBE++
		} else if c1 == 0x00 && (0x03 < c0 && c0 < 0x20) {
			d.scoreLE++
		}
		if surr1 <= c0 && c0 < surr2 {
			if surr2 <= c1 && c1 < surr3 {
				//BE
				d.scoreBE++
				if d.scoreLE > 0 {
					err = ErrInvalidEncoding
					break loop
				}
			} else {
				err = ErrInvalidEncoding
				break loop
			}
		} else if surr1 <= c1 && c1 < surr2 {
			if surr2 <= c0 && c0 < surr3 {
				//LE
				d.scoreLE++
				if d.scoreLE > 0 {
					err = ErrInvalidEncoding
					break loop
				}
			} else {
				err = ErrInvalidEncoding
				break loop
			}
		}
	}

	if atEOF && err == transform.ErrShortSrc {
		err = ErrInvalidEncoding
	}

	if d.scoreLE != 0 {
		score = -d.scoreLE
	} else if d.scoreBE != 0 {
		score = +d.scoreBE
	}

	return nSrc, err, score
}

func (c *utf16Encoding) getDecoder() transform.Transformer {
	if c.decoder == nil {
		c.decoder = unicode.UTF16(c.endian, c.bom).NewDecoder()
	} else {
		c.decoder.Reset()
	}

	return c.decoder
}

func (c *utf16Encoding) Decode(b []byte) (string, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(b), c.getDecoder()))
	if err != nil {
		return "", err
	}
	return string(ret), err
}


func (c *utf16Encoding) getEncoder() transform.Transformer {
	if c.encoder == nil {
		c.encoder = unicode.UTF16(c.endian, c.bom).NewEncoder()
	} else {
		c.encoder.Reset()
	}

	return c.encoder
}

func (c *utf16Encoding) Encode(s string) ([]byte, error) {
	ret, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(s), c.getEncoder()))
	if err != nil {
		return nil, err
	}
	return ret, err
}

type utf16splitter struct {
	cr      bool
	remains []byte

	endian unicode.Endianness
}

func (sp *utf16splitter) reset() {
	sp.cr = false
	sp.remains = []byte{}
}

func (sp *utf16splitter) Split(src []byte, atEnd bool, lines *list.List) {

	start := 0
	src = append(sp.remains, src...)
	start = len(sp.remains)
	if sp.cr {
		start -= 2
	}
	sp.reset()

	n := len(src)
	nHead := 0

	if sp.endian == unicode.BigEndian {
		for i := start; i < n; i += 2 {
			c0 := src[i]
			c1 := src[i+1]
			if c0 == 0x00 && c1 == CR {
				if n <= i+4 {
					sp.cr = true
					break
				}

				c2 := src[i+2]
				c3 := src[i+3]
				if c2 == 0x00 && c3 == LE {
					i += 4
				}
				lines.PushBack(src[nHead : i+2])
				nHead = i + 2
			} else if c0 == 0x00 && c1 == LE {
				lines.PushBack(src[nHead : i+2])
				nHead = i + 2
			}
		}
	} else {
		for i := start; i < n; i += 2 {
			c1 := src[i]
			c0 := src[i+1]
			if c0 == 0x00 && c1 == CR {
				if n <= i+4 {
					sp.cr = true
					break
				}

				c3 := src[i+2]
				c2 := src[i+3]
				if c2 == 0x00 && c3 == LE {
					i += 4
				}
				lines.PushBack(src[nHead : i+2])
				nHead = i + 2
			} else if c0 == 0x00 && c1 == LE {
				lines.PushBack(src[nHead : i+2])
				nHead = i + 2
			}
		}
	}

	if nHead <= n-1 {
		sp.remains = src[nHead:n]
	}
}
