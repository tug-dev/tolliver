package binary

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/internal/common"
)

type Reader struct {
	*bufio.Reader
}

func NewReader(conn net.Conn) *Reader {
	return &Reader{bufio.NewReader(conn)}
}

func (r *Reader) ReadAll(lens []uint32, destinations ...any) error {
	var p = 0

	for _, val := range destinations {
		switch v := val.(type) {
		case *byte:
			res, err := r.ReadByte()
			if err != nil {
				return err
			}

			*v = res

		case *uint64:
			res, err := r.ReadUint64()
			if err != nil {
				return err
			}

			*v = res

		case *uint32:
			res, err := r.ReadUint32()
			if err != nil {
				return err
			}

			*v = res

		case *string:
			res, err := r.ReadString(lens[p])
			if err != nil {
				return err
			}
			p++

			*v = res

		case *uuid.UUID:
			res, err := r.ReadUUID()
			if err != nil {
				return err
			}

			*v = res

		case *[]common.SubcriptionInfo:
			err := r.ReadSubs(v)
			if err != nil {
				return err
			}

		default:
			panic("Unsupported reader type")
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

func (r *Reader) ReadString(length uint32) (string, error) {
	b := make([]byte, length)
	_, err := io.ReadFull(r, b)

	return string(b), err
}

func (r *Reader) FillBuf(buf []byte) error {
	_, err := io.ReadFull(r, buf)
	return err
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

func (r *Reader) ReadSubs(dest *[]common.SubcriptionInfo) error {
	num, err := r.ReadUint32()
	if err != nil {
		return err
	}

	if &dest == nil {
		*dest = make([]common.SubcriptionInfo, num)
	}

	for i := range num {
		chanLen, err := r.ReadUint32()
		if err != nil {
			return err
		}
		keyLen, err := r.ReadUint32()
		if err != nil {
			return err
		}

		channel, err := r.ReadString(chanLen)
		if err != nil {
			return err
		}
		key, err := r.ReadString(keyLen)
		if err != nil {
			return err
		}

		(*dest)[i] = common.SubcriptionInfo{Channel: channel, Key: key}
	}

	return nil
}
