package types

import (
	"io"
)

type Byte byte

func (b Byte) Encode() []byte {
	return []byte{byte(b)}
}

func (b *Byte) Decode(r io.ByteReader) error {

	v, err := r.ReadByte()
	if err != nil {
		return err
	}

	*b = Byte(v)
	return nil

}
