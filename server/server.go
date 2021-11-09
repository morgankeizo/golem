package server

type ServerState int

// Server state values
const (
	Stopped ServerState = iota
	Starting
	Running
	Stopping
)

// StopCommand is the Minecraft server stop command.
const StopCommand = "stop"

// Server is the interface that defines a server manager.
type Server interface {
	Start() error
	Stop() error
	Execute(command string) (string, error)
	State() ServerState
}
