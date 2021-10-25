package types

import (
	"io"
)

type UnsignedShort uint16

func (u *UnsignedShort) Decode(r io.ByteReader) error {

	byte1, err := r.ReadByte()
	if err != nil {
		return err
	}

	byte2, err := r.ReadByte()
	if err != nil {
		return err
	}

	*u = UnsignedShort(int16(byte1)<<8 | int16(byte2))
	return nil

}
