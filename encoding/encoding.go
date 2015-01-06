package encoding

import (
	"container/list"
	"errors"
	"fmt"
	"golang.org/x/text/transform"
)

// Line Separator
const (
	LE byte = 0x0A
	CR byte = 0x0D
)

var ErrInvalidEncoding = errors.New("invalid encoding")

type Encoding interface {
	Splitter
	Decoder
	Encoder
	fmt.Stringer
}


type EncodingSearcher interface {
	EncodingSearch(p []byte, atEOF bool) (nSrc int, err error, score int)
}

func CheckEncoding(ls *list.List, es EncodingSearcher) (yes bool, score int) {
	e := ls.Front()
	var nSrc, s int
	var err error
	var prv []byte
	for {
		b := e.Value.([]byte)
		if prv != nil {
			b = append(prv, b...)
		}
		next := e.Next()
		if next == nil {
			nSrc, err, s = es.EncodingSearch(b, true)
			score += s
			break
		} else {
			nSrc, err, s = es.EncodingSearch(b, false)
			score += s
			if err == ErrInvalidEncoding {
				break
			} else if err == transform.ErrShortSrc {
				prv = b[nSrc:]
			} else {
				prv = nil
			}
		}

		e = next
	}

	return err == nil, score
}

type Splitter interface {
	Split(src []byte, atEnd bool, lines *list.List)
}

type splitter struct {
	cr      bool
	remains []byte
}

func (sp *splitter) reset() {
	sp.cr = false
	sp.remains = []byte{}
}

func (sp *splitter) Split(src []byte, atEnd bool, lines *list.List) {

	//	fmt.Printf("****** [%x]\n", src[0])
	//	if len(sp.remains) > 0 {
	//		fmt.Printf("****** [%x]\n", sp.remains[0])
	//		}

	start := 0
	src = append(sp.remains, src...)
	start = len(sp.remains)
	if sp.cr {
		start--
	}
	sp.reset()

	n := len(src)
	nHead := 0
	//	fmt.Printf("***** %d %d [%x]\n", start, n, src[start])
loop:
	for i := start; i < n; i++ {
		c0 := src[i]
		//		fmt.Printf("%x \n", c0)
		if c0 == CR {
			//			println("!")
			if n <= i+1 {
				//				println("****", n, i+1)
				sp.cr = true
				break loop
			}

			c1 := src[i+1]
			if c1 == LE {
				i += 1
			}
			lines.PushBack(src[nHead : i+1])
			//			println("**", n, string(src[nHead:i+1]))
			nHead = i + 1
		} else if c0 == LE {
			//			println("!!")
			lines.PushBack(src[nHead : i+1])
			//			println("***", n, string(src[nHead:i+1]))
			nHead = i + 1
		}
	}

	if nHead <= n-1 {
		if atEnd {
			lines.PushBack(src[nHead:n])
		} else {
			//			fmt.Printf("* %d %d [%x]\n", nHead, n, src[nHead])
			sp.remains = src[nHead:n]
		}
	}
}

type Decoder interface {
	Decode(b []byte) (string, error)
}

type Encoder interface {
	Encode(s string) ([]byte, error)
}
