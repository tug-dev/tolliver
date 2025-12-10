package tolliver

import (
	"bufio"
	"encoding/binary"
	"io"
)

type Reader struct {
	*bufio.Reader
}

func (r *Reader) ReadUint64() (uint64, error) {
	s := make([]byte, 8)
	_, err := io.ReadFull(r, s)

	return binary.BigEndian.Uint64(s), err
}

func (r *Reader) ReadUint32() (uint32, error) {
	s := make([]byte, 4)
	_, err := io.ReadFull(r, s)

	return binary.BigEndian.Uint32(s), err
}
