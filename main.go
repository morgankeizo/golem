package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	proxyPkg "golem/proxy"
	serverPkg "golem/server"
	"golem/server/process"
)

func main() {

	var proxyAddr string
	var serverAddr string
	var serverStart string
	var serverDirectory string
	var stopTimeout int
	var versionName string
	var versionProtocol int
	var playersMax int
	var debug bool

	// Define flags
	flag.StringVar(&proxyAddr, "proxyAddr", ":25565",
		"Proxy server address")
	flag.StringVar(&serverAddr, "serverAddr", ":25566",
		"Minecraft server address")
	flag.StringVar(&serverStart, "serverStart", "",
		"Minecraft start command. Empty disables autostart/stop")
	flag.StringVar(&serverDirectory, "serverDirectory", "",
		"Minecraft server working directory")
	flag.IntVar(&stopTimeout, "stopTimeout", 60,
		"Wait period to stop server after last disconnect (seconds)")
	flag.StringVar(&versionName, "versionName", "1.17.1",
		"Minecraft version name")
	flag.IntVar(&versionProtocol, "versionProtocol", 756,
		"Minecraft protocol version")
	flag.IntVar(&playersMax, "playersMax", 20,
		"Maximum number of players (to display in status message)")
	flag.BoolVar(&debug, "debug", false,
		"Log all traffic")
	flag.Parse()

	// Create server depending on if server start command was given
	var server serverPkg.Server
	var timeDuration *time.Duration
	if serverStart == "" {
		server = serverPkg.NewDummyServer()
	} else {
		server = process.NewProcessServer(
			newLogger("[server] "),
			strings.Fields(serverStart),
			serverDirectory,
		)

		// Make time duration reference
		d := time.Duration(stopTimeout) * time.Second
		timeDuration = &d
	}

	// Make optional packet logger
	var protocolLogger *log.Logger
	if debug {
		protocolLogger = newLogger("[protocol] ")
	}

	// Make proxy
	proxy := proxyPkg.NewProxy(
		newLogger("[proxy] "),
		proxyAddr,
		serverAddr,
		timeDuration,
		server,
		protocolLogger,
		versionName,
		versionProtocol,
		playersMax,
	)

	// Listen for SIGINT or SIGTERM and safely exit
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		server.Stop()
		os.Exit(1)
	}()

	// Run proxy
	err := proxy.Run()
	if err != nil {
		fmt.Println(err)
	}

}

// newLogger makes a new logger to stdout.
func newLogger(prefix string) *log.Logger {
	return log.New(os.Stdout, prefix, 0)
}
