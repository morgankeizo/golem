package protocol

const (
	// Serverbound
	StatusRequestPacketID = byte(0)
	StatusPingPacketID    = byte(1)

	// Clientbound
	StatusResponsePacketID = byte(0)
)

type StatusRequestPacket struct{}

type StatusPingPacket struct{}

type StatusResponsePacket struct {
	StatusResponse string `protocol:"String"`
}

type ServerStatus struct {
	Version     Version     `json:"version"`
	Players     Players     `json:"players"`
	Description Description `json:"description"`
	Favicon     string      `json:"favicon,omitempty"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type Players struct {
	Max    int            `json:"max"`
	Online int            `json:"online"`
	Sample []PlayerSample `json:"sample,omitempty"`
}

type PlayerSample struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type Description struct {
	Text string `json:"text"`
}

type ServerText struct {
	Text string `json:"text"`
}
