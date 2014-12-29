package text

import (
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

func (c *Text) WriteTo(out io.Writer) {
	itr := c.Iterator()
	for itr.HasNext() {
		str := itr.Next()
		out.Write([]byte(str))
	}
}
