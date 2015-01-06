package text

import (
	"bufio"
	"container/list"
	"github.com/zackys/go.p/encoding"
	"github.com/zackys/go.p/file"
	"io"
)

type Text struct {
	ls *list.List

	encoding encoding.Encoding
}

func New(encoding encoding.Encoding) *Text {
	return &Text{
		list.New(),
		encoding,
	}
}

type Iterator struct {
	next *list.Element
}

func (c *Text) Iterator() Iterator {
	return Iterator{
		next: c.ls.Front(),
	}
}

func (itr Iterator) HasNext() bool {
	return itr.next != nil
}

func (itr *Iterator) Next() string {
	ret := itr.next.Value.(string)
	itr.next = itr.next.Next()
	return ret
}

func (c *Text) ReadFrom(in *file.Bytes) {
	ls := list.New()
	itr := in.Iterator()
	for itr.HasNext() {
		b := itr.Next()
		c.encoding.Split(b, !itr.HasNext(), ls)
	}
	for e := ls.Front(); e != nil; e = e.Next() {
		str, _ := c.encoding.Decode(e.Value.([]byte))
		c.ls.PushBack(str)
	}
}

func (c *Text) Transform(t ...Transformer) error {
	var err error
	newlst := list.New()
	itr := c.Iterator()
	for itr.HasNext() {
		s := itr.Next()
		for _, t0 := range t {
			s, err = t0.Transform(s)
			if err != nil {
				return err
			}
		}
		newlst.PushBack(s)
	}

	c.ls = newlst
	return nil
}

func (c *Text) WriteTo(out io.Writer, enc encoding.Encoder) error {
	writer := bufio.NewWriter(out)
	itr := c.Iterator()
	for itr.HasNext() {
		b, err := enc.Encode(itr.Next())
		if err != nil {
			return err
		} else {
			writer.Write(b)
		}
	}
	return writer.Flush()
}
