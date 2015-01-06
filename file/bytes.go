package file

import (
	"bufio"
	"container/list"
	"github.com/zackys/go.p/encoding"
	"golang.org/x/text/transform"
	"io"
	"log"
	"os"
)

type Bytes struct {
	ls *list.List
}

func NewBytes() *Bytes {
	return &Bytes{
		list.New(),
	}
}

type Iterator struct {
	next *list.Element
}

func (c *Bytes) Iterator() Iterator {
	return Iterator{
		next: c.ls.Front(),
	}
}

func (itr Iterator) HasNext() bool {
	return itr.next != nil
}

func (itr *Iterator) Next() []byte {
	ret := itr.next.Value.([]byte)
	itr.next = itr.next.Next()
	return ret
}

func (c *Bytes) ReadFrom(in *os.File) error {
	r := bufio.NewReader(in)
	for {
		b := make([]byte, 2)
		n, err := r.Read(b)
		if n == 0 && err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		c.ls.PushBack(b[:n])
	}

	return nil
}

func (c *Bytes) WriteTo(out *os.File) error {

	return nil
}

func checkEncoding(c *Bytes, es encoding.EncodingSearcher) (yes bool, score int) {
	var nSrc, s int
	var err error
	var prv []byte
	itr := c.Iterator()
	for itr.HasNext() {
		b := itr.Next()
		if prv != nil {
			b = append(prv, b...)
		}

		nSrc, err, s = es.EncodingSearch(b, !itr.HasNext())
		score += s
		if err == encoding.ErrInvalidEncoding {
			break
		} else if err == transform.ErrShortSrc {
			prv = b[nSrc:]
		} else {
			prv = nil
		}
	}

	return err == nil, score
}

func (c *Bytes) SearchEncoding() encoding.Encoding {
	var enc encoding.Encoding

	yes, score := checkEncoding(c, encoding.ISO2022JP)
	if yes {
		if score > 0 {
			//ISO2022確定
			debug("ISO2022")
			enc = encoding.ISO2022JP
		} else {
			//ASCII確定
			debug("ASCII")
			enc = encoding.ASCII
		}
	} else {
		yes, score := checkEncoding(c, encoding.UTF8)
		if yes {
			//UTF8確定
			debug("UTF8")
			enc = encoding.UTF8
		} else {
			sjis, scoreSjis := checkEncoding(c, encoding.ShiftJIS)
			if sjis {
				//ShiftJIS確定
				debug("maybe ShiftJIS")
			}
			euc, scoreEuc := checkEncoding(c, encoding.EUCJP)
			if euc {
				if !sjis {
					debug("EUCJP", scoreEuc, scoreSjis)
					enc = encoding.EUCJP
				} else if scoreSjis < scoreEuc {
					debug("EUCJP", scoreEuc, scoreSjis)
					enc = encoding.EUCJP
				} else {
					debug("*ShiftJIS", scoreSjis, scoreEuc)
					enc = encoding.ShiftJIS
				}
			} else {
				if sjis {
					debug("**ShiftJIS", scoreSjis, scoreEuc)
					enc = encoding.ShiftJIS
				} else {
					yes, score = checkEncoding(c, encoding.UTF16)
					if yes {
						if score > 0 {
							debug("UTF16BE")
							enc = encoding.UTF16BE
						} else {
							debug("UTF16LE")
							enc = encoding.UTF16LE
						}
					} else {
						debug("none")
					}
				}
			}
		}
	}

	return enc
}

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}
