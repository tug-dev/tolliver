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

func (r *Reader) ReadAll(lens []uint32, destinations ...any) error {
	var p = 0

	for _, val := range destinations {
		switch v := val.(type) {
		case *uint64:
			val, err := r.ReadUint64()
			if err != nil {
				return err
			}

			*v = val

		case *uint32:
			val, err := r.ReadUint32()
			if err != nil {
				return err
			}

			*v = val

		case *string:
			val, err := r.ReadString(lens[p])
			if err != nil {
				return err
			}
			p++

			*v = val

		case *uuid.UUID:
			val, err := r.ReadUUID()
			if err != nil {
				return err
			}

			*v = val
		}
	}

	return nil
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
