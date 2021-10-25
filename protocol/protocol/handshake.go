package protocol

const HandshakePacketID = byte(0)

const (
	NextStateStatusRequest = 1
	NextStateLoginRequest  = 2
)

type HandshakePacket struct {
	Data            []byte `protocol:"_data"`
	ProtocolVersion int    `protocol:"VarInt"`
	ServerAddress   string `protocol:"String"`
	ServerPort      int    `protocol:"UnsignedShort"`
	NextState       int    `protocol:"VarInt"`
}
