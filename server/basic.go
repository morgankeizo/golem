package server

type BasicServer struct{}

// NewBasicServer returns a new basic server.
func NewBasicServer() *BasicServer {
	return &BasicServer{}
}

// Start implements Server.
func (s *BasicServer) Start() error {
	return nil
}

// Stop implements Server.
func (s *BasicServer) Stop() error {
	return nil
}

// Execute implements Server.
func (s *BasicServer) Execute(command string) (string, error) {
	return "", nil
}

// State implements Server.
func (s *BasicServer) State() ServerState {
	return Running
}
