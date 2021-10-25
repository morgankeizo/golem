package proxy

const (
	statusStarting = "[starting]"
	statusStopping = "[stopping]"
	statusRunning  = "[running]"
	statusStopped  = "[stopped]"
)

const (
	serverStopped         = "server is stopped"
	serverStarting        = "server is starting..."
	serverStopping        = "server is stopping..."
	serverStartInitiated  = "server start initiated"
	serverStartFailed     = "server start failed"
	serverConnectFailed   = "server connect failed"
	serverHandshakeFailed = "server handshake failed"
)
