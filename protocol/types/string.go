package types

import (
	"io"
)

type String string

func (s String) Encode() []byte {

	b := []byte(s)
	return append(
		VarInt(len(b)).Encode(),
		b...,
	)

}

func (s *String) Decode(r io.ByteReader) error {

	var length VarInt
	err := length.Decode(r)
	if err != nil {
		return err
	}

	n := int(length)
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i], err = r.ReadByte()
		if err != nil {
			return err
		}
	}

	*s = String(bytes)
	return nil

}
