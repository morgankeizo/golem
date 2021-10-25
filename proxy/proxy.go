package proxy

import (
	"io"
	"log"
	"net"
	"time"

	"golem/protocol"
	protocolDefinitions "golem/protocol/protocol"
	serverPkg "golem/server"
)

// A Proxy proxies a Minecraft server.
type Proxy struct {
	logger         *log.Logger
	protocolLogger *log.Logger
	server         serverPkg.Server
	proxyAddr      string
	serverAddr     string

	stopDuration *time.Duration // nil disables autostart/stop
	stopTimer    *time.Timer

	players map[string]bool // set of usernames

	versionName     string
	versionProtocol int
	playersMax      int
}

// NewProxy returns a new Proxy.
//
// Autostart/stop is disabled when stopDuration is nil.
// Optional packet logging is disabled when protocolLogger is nil.
func NewProxy(
	logger *log.Logger,
	proxyAddr string,
	serverAddr string,
	stopDuration *time.Duration,
	server serverPkg.Server,
	protocolLogger *log.Logger,
	versionName string,
	versionProtocol int,
	playersMax int,
) *Proxy {
	p := Proxy{}
	p.logger = logger
	p.proxyAddr = proxyAddr
	p.serverAddr = serverAddr
	p.stopDuration = stopDuration
	p.server = server
	p.players = make(map[string]bool)
	p.protocolLogger = protocolLogger
	p.versionName = versionName
	p.versionProtocol = versionProtocol
	p.playersMax = playersMax
	return &p
}

// Run starts a proxy listen loop.
func (p *Proxy) Run() error {

	listener, err := net.Listen("tcp", p.proxyAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			p.logger.Printf("error accepting packet: %s\n", err)
			continue
		}

		go p.handleConnection(protocol.NewClientConn(conn, p.protocolLogger))
	}

	return nil

}

// handleConnection handles an incoming connection.
func (p *Proxy) handleConnection(conn *protocol.ClientConn) {

	// Read handshake packet
	handshakePacket, err := conn.ReadHandshakePacket()
	if err != nil {
		p.logger.Printf("error reading handshape packet: %s\n", err)
		return
	}

	// Handle depending on handshake next state
	switch handshakePacket.NextState {
	case protocolDefinitions.NextStateStatusRequest:

		// Read status request packet
		_, err = conn.ReadStatusRequestPacket()
		if err != nil {
			p.logger.Printf("error reading status request packet: %s\n", err)
			return
		}

		// Write status message depending on server state
		var statusMessage string
		switch p.server.State() {
		case serverPkg.Starting:
			statusMessage = statusStarting
		case serverPkg.Stopped:
			statusMessage = statusStopped
		case serverPkg.Running:
			statusMessage = statusRunning
		case serverPkg.Stopping:
			statusMessage = statusStopping
		}
		err = conn.WriteMessageStatus(
			statusMessage,
			p.versionName,
			p.versionProtocol,
			len(p.players),
			p.playersMax,
		)
		if err != nil {
			p.logger.Printf("error sending message: %s\n", err)
			return
		}

		// Read and respond to ping packet
		err = conn.ReadAndRespondPing()
		if err != nil {
			p.logger.Printf("error handling ping: %s\n", err)
			return
		}

	case protocolDefinitions.NextStateLoginRequest:

		// Write text message depending on server state
		// Continue only when state is Running
		switch p.server.State() {
		case serverPkg.Starting:
			err = conn.WriteMessageText(serverStarting)
			if err != nil {
				p.logger.Printf("error sending message: %s\n", err)
			}
			return
		case serverPkg.Stopping:
			err = conn.WriteMessageText(serverStopping)
			if err != nil {
				p.logger.Printf("error sending message: %s\n", err)
			}
			return
		case serverPkg.Stopped:

			// Start server is autostart/stop enabled
			if p.stopDuration != nil {
				p.logger.Println("starting server")
				err = p.server.Start()
				if err != nil {
					err = conn.WriteMessageText(serverStartFailed)
				} else {
					err = conn.WriteMessageText(serverStartInitiated)
				}
			} else {
				err = conn.WriteMessageText(serverStopped)
			}
			if err != nil {
				p.logger.Printf("error sending message: %s\n", err)
				return
			}

			return
		}

		// Read login start packet
		loginPacket, err := conn.ReadLoginStartPacket()
		if err != nil {
			p.logger.Printf("error reading login start packet: %s\n", err)
			return
		}

		// Connect to server
		serverConn, err := net.Dial("tcp", p.serverAddr)
		if err != nil {
			p.logger.Printf("error connecting to server: %s\n", err)
			conn.WriteMessageText(serverConnectFailed)
			return
		}
		defer serverConn.Close()

		// Catch up server connection
		_, err = serverConn.Write(handshakePacket.Data)
		if err != nil {
			p.logger.Printf("error writing to server: %s\n", err)
			return
		}
		_, err = serverConn.Write(loginPacket.Data)
		if err != nil {
			p.logger.Printf("error writing to server: %s\n", err)
			return
		}

		// Player connected
		username := loginPacket.Username
		p.logger.Printf("player connected: %s\n", username)
		p.players[username] = true
		if p.stopDuration != nil && p.stopTimer != nil {
			p.logger.Println("reseting stop timer")
			p.stopTimer.Stop()
			p.stopTimer = nil
		}

		// Pipe connections in both directions
		// Ensure pipes close together
		stop := false
		go p.pipe(serverConn, conn, &stop)
		p.pipe(conn, serverConn, &stop)

		// Player disconnected
		p.logger.Printf("player disconnected: %s\n", username)
		delete(p.players, username)
		if p.stopDuration != nil && len(p.players) == 0 {
			p.logger.Println("starting stop timer")
			p.stopTimer = time.AfterFunc(*p.stopDuration, func() {
				p.server.Stop()
				p.stopTimer = nil
			})
		}

	}

}

// pipe wraps the pipe implementation to catch errors.
func (p *Proxy) pipe(src io.ReadCloser, dst io.WriteCloser, stop *bool) {
	err := pipe(src, dst, stop)
	if err != nil {
		p.logger.Printf("error forwarding connection: %s\n", err)
	}
}
