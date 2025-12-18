package binary

import (
	"encoding/binary"

	"github.com/google/uuid"
	"github.com/tug-dev/tolliver/go/common"
)

type Writer struct {
	ptr  uint32
	data []byte
}

func NewWriter() *Writer {
	out := Writer{
		data: make([]byte, 0, 32),
	}

	return &out
}

func (w *Writer) WriteAll(sources ...any) {
	for _, v := range sources {
		switch val := v.(type) {
		case byte:
			w.WriteByte(val)

		case uint32:
			w.WriteUint32(val)

		case uint64:
			w.WriteUint64(val)

		case uuid.UUID:
			w.WriteUUID(val)

		case []common.SubcriptionInfo:
			w.WriteSubscriptions(val)
		}
	}
}

func (w *Writer) Join() []byte {
	return w.data
}

func (w *Writer) WriteByte(b byte) {
	w.data = append(w.data, b)
}

func (w *Writer) WriteUint64(n uint64) {
	w.data = binary.BigEndian.AppendUint64(w.data, n)
}

func (w *Writer) WriteUint32(n uint32) {
	w.data = binary.BigEndian.AppendUint32(w.data, n)
}

func (w *Writer) WriteUUID(id uuid.UUID) {
	w.data = append(w.data, id[:]...)
}

func (w *Writer) WriteSubscriptions(subs []common.SubcriptionInfo) {
	for _, v := range subs {
		w.WriteUint32(uint32(len(v.Channel)))
		w.WriteUint32(uint32(len(v.Key)))
		w.data = append(w.data, []byte(v.Channel)...)
		w.data = append(w.data, []byte(v.Key)...)
	}
}
