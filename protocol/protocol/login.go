package protocol

const LoginStartPacketID = byte(0)

type LoginStartPacket struct {
	Data     []byte `protocol:"_data"`
	Username string `protocol:"String"`
}
