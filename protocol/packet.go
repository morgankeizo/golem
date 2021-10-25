package protocol

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"golem/protocol/types"
)

const bufferSize = 1024

const tagName = "protocol"

// readPacket reads and decodes a packet into a struct.
// Returns an error if:
// - the expected packet id differs from the packet id read,
// - the struct contained an invalid type, or
// - an error occured during read.
func (c *ClientConn) readPacket(p interface{}, id byte) error {

	r, data, err := readPacket(c, id)
	if err != nil {
		return err
	}

	return decodePacket(p, r, data)

}

// writePacket encodes and writes a packet from a struct.
func (c *ClientConn) writePacket(p interface{}, id byte) error {

	data, err := encodePacket(p)
	if err != nil {
		return err
	}

	return writePacket(c, id, data)

}

// readPacket reads a packet from a reader, returning:
// - a ByteReader of the data section of the packet,
// - the bytes of the entire packet, and
// - if an error occured during read or
// - the expected packet id differs from the packet id read.
func readPacket(
	r interface {
		io.Reader
		io.ByteReader
		BytesReader
	},
	id byte,
) (io.ByteReader, []byte, error) {

	var err error
	var data []byte
	var br io.ByteReader
	var packetLength types.VarInt
	var packetID types.Byte

	err = func() error {

		err = packetLength.Decode(r)
		if err != nil {
			return err
		}

		data, err = r.ReadBytes(int(packetLength))
		if err != nil {
			return err
		}
		br = bytes.NewReader(data)

		err = packetID.Decode(br)
		if err != nil {
			return err
		}

		if byte(packetID) != id {
			return fmt.Errorf(
				"expected packet id %x but got %x",
				id,
				byte(packetID),
			)
		}

		return nil

	}()
	if err != nil {
		return nil, []byte{}, err
	}

	data = append(packetLength.Encode(), data...)
	return br, data, nil

}

// writePacket writes a packet to a writer.
func writePacket(w io.Writer, id byte, data []byte) error {

	packetIDBytes := types.Byte(id).Encode()
	packetLength := types.VarInt(len(packetIDBytes) + len(data))

	_, err := w.Write(packetLength.Encode())
	if err != nil {
		return nil
	}

	_, err = w.Write(packetIDBytes)
	if err != nil {
		return nil
	}

	_, err = w.Write(data)
	return err

}

// decodePacket decodes a packet from a tagged struct using reflection.
func decodePacket(p interface{}, r io.ByteReader, data []byte) error {

	packetValue := reflect.ValueOf(p).Elem()
	packetType := packetValue.Type()

	for i := 0; i < packetType.NumField(); i++ {

		var value reflect.Value

		valueField := packetValue.Field(i)
		typeField := packetType.Field(i)
		tag := typeField.Tag.Get(tagName)

		switch tag {
		case "_data":
			value = reflect.ValueOf(data)
		case "Byte":
			var v types.Byte
			v.Decode(r)
			value = reflect.ValueOf(v)
		case "String":
			var v types.String
			v.Decode(r)
			value = reflect.ValueOf(v)
		case "UnsignedShort":
			var v types.UnsignedShort
			v.Decode(r)
			value = reflect.ValueOf(v)
		case "VarInt":
			var v types.VarInt
			v.Decode(r)
			value = reflect.ValueOf(v)
		default:
			return fmt.Errorf("unknown protocol type: %s", tag)
		}

		valueField.Set(value.Convert(typeField.Type))

	}

	return nil

}

// encodePacket encodes a packet to a tagged struct using reflection.
func encodePacket(p interface{}) ([]byte, error) {

	var buffer bytes.Buffer

	packetValue := reflect.ValueOf(p).Elem()
	packetType := packetValue.Type()

	for i := 0; i < packetType.NumField(); i++ {

		var encoder interface {
			Encode() []byte
		}

		valueField := packetValue.Field(i)
		typeField := packetType.Field(i)
		tag := typeField.Tag.Get(tagName)

		switch tag {
		case "_data":
			continue
		case "Byte":
			encoder = types.Byte(valueField.Int())
		case "String":
			encoder = types.String(valueField.String())
		// case "UnsignedShort":
		// 	encoder = types.UnsignedShort(valueField.Int())
		case "VarInt":
			encoder = types.VarInt(valueField.Int())
		default:
			return []byte{}, fmt.Errorf("unknown protocol type: %s", tag)
		}

		_, err := buffer.Write(encoder.Encode())
		if err != nil {
			return []byte{}, err
		}

	}

	return buffer.Bytes(), nil

}
