package types

import (
	"fmt"
	"io"
)

type VarInt int32

func (v VarInt) Encode() []byte {

	value := uint32(v)
	bytes := []byte{}

	for {

		if value&0xFFFFFF80 == 0 {
			bytes = append(bytes, byte(value))
			break
		}

		bytes = append(bytes, byte(value&0x7F|0x80))
		value >>= 7

	}

	return bytes

}

func (v *VarInt) Decode(r io.ByteReader) error {

	var value uint32

	for i := 0; ; i++ {

		current, err := r.ReadByte()
		if err != nil {
			return err
		}

		value |= uint32(current&0x7F) << uint32(7*i)

		if i >= 5 {
			return fmt.Errorf("VarInt is too big")
		}

		if current&0x80 == 0 {
			break
		}

	}

	*v = VarInt(value)
	return nil

}
