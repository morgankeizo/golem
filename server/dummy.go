package server

type DummyServer struct{}

// NewDummyServer returns a new dummy server.
func NewDummyServer() *DummyServer {
	return &DummyServer{}
}

// Start implements Server.
func (s *DummyServer) Start() error {
	return nil
}

// Stop implements Server.
func (s *DummyServer) Stop() error {
	return nil
}

// Execute implements Server.
func (s *DummyServer) Execute(command string) (string, error) {
	return "", nil
}

// State implements Server.
func (s *DummyServer) State() ServerState {
	return Running
}
