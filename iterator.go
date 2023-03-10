package iterator

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrPositiveSize is returned when a negative page size is passed to NewPager.
	ErrPositiveSize = errors.New("iterator: page size must be positive")
	// ErrBufferNotEmpty is returned when NextPage is called with a non-empty buffer.
	ErrBufferNotEmpty = errors.New("pager: must call NextPage with an empty buffer")
	// ErrNilNextPage is returned when NextPage is called with a nil pointer.
	ErrNilNextPage = errors.New("pager: nil passed to Pager.NextPage")
)

// Iterator represents the state of an iterator.
type Iterator[V any] struct {
	MaxSize int
	Token   string

	err         error
	atLast      bool
	retrieve    func(pageSize int, pageToken string) (nextPageToken string, err error)
	retrieveBuf func() V

	bufLen func() int

	nextCalled     bool
	nextPageCalled bool
}

// Pageable is implemented by iterators that support paging.
type Pageable[V any] interface {
	Iterator() *Iterator[V]
}

// Pager represents the state of a pager
type Pager[V any] struct {
	iterator *Iterator[V]
	size     int
}

// NewIterator returns a new Iterator that will call retrieve to retrieve
func NewIterator[V any](retrieve func(int, string) (string, error), bufLen func() int, retrieveBuf func() V) (*Iterator[V], func() error) {
	i := &Iterator[V]{
		retrieve:    retrieve,
		retrieveBuf: retrieveBuf,
		bufLen:      bufLen,
	}

	return i, i.next
}

// Len returns the number of items in the buffer.
func (i *Iterator[V]) Len() int {
	return i.bufLen()
}

func (i *Iterator[V]) next() error {
	i.nextCalled = true

	if i.err != nil {
		return i.err
	}

	for i.bufLen() == 0 && !i.atLast {
		if err := i.buffer(i.MaxSize); err != nil {
			i.err = err
			return i.err
		}
		if i.Token == "" {
			i.atLast = true
		}
	}

	if i.bufLen() == 0 {
		i.err = nil
	}

	return i.err
}

func (i *Iterator[V]) buffer(size int) error {
	token, err := i.retrieve(size, i.Token)
	if err != nil {
		i.retrieveBuf()
		return err
	}

	i.Token = token

	return nil
}

// NewPager returns a new Pager that will call retrieve to retrieve
func NewPager[V any](iter Pageable[V], size int, token string) *Pager[V] {
	p := &Pager[V]{iter.Iterator(), size}
	p.iterator.Token = token

	if size < 0 {
		p.iterator.err = ErrPositiveSize
	}

	return p
}

// NextPage retrieves the next page of results and appends them to the buffer.
func (p *Pager[V]) NextPage(ptr *V) (string, error) {
	p.iterator.nextPageCalled = true

	if p.iterator.err != nil {
		return "", p.iterator.err
	}

	if p.iterator.bufLen() > 0 {
		return "", ErrBufferNotEmpty
	}

	bufType := reflect.PtrTo(reflect.ValueOf(p.iterator.retrieveBuf()).Type())
	if ptr == nil {
		return "", ErrNilNextPage
	}

	ptrValue := reflect.ValueOf(ptr)
	if ptrValue.Type() != bufType {
		return "", fmt.Errorf("pager: next should be of type %s, got %T", bufType, ptr)
	}

	for p.iterator.bufLen() < p.size {
		if err := p.iterator.buffer(p.size - p.iterator.bufLen()); err != nil {
			return "", p.iterator.err
		}

		if p.iterator.Token == "" {
			break
		}
	}

	e := ptrValue.Elem()
	e.Set(reflect.AppendSlice(e, reflect.ValueOf(p.iterator.retrieveBuf())))

	return p.iterator.Token, nil
}
