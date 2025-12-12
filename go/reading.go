package tolliver

import (
	"bufio"
	"encoding/binary"
	"io"

	"github.com/google/uuid"
)

type Reader struct {
	*bufio.Reader
}

func (r *Reader) ReadUint64() (uint64, error) {
	b := make([]byte, 8)
	_, err := io.ReadFull(r, b)

	return binary.BigEndian.Uint64(b), err
}

func (r *Reader) ReadUint32() (uint32, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)

	return binary.BigEndian.Uint32(b), err
}

func (r *Reader) ReadString(len uint32) (string, error) {
	b := make([]byte, len)
	_, err := io.ReadFull(r, b)

	return string(b), err
}

func (r *Reader) ReadUUID() (uuid.UUID, error) {
	b := make([]byte, 16)
	_, err := io.ReadFull(r, b)

	if err != nil {
		return uuid.UUID{}, err
	}

	id, _ := uuid.FromBytes(b)

	return id, nil
}
