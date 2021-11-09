package process

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	serverPkg "golem/server"
)

// A ProcessServer implements server.Server by supervising a process.
type ProcessServer struct {
	state serverPkg.ServerState

	logger          *log.Logger
	serverStartArgs []string
	serverDirectory string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	wg     sync.WaitGroup
	lines  chan string
	exited chan bool
}

// NewProcessServer returns a new ProcessServer.
func NewProcessServer(
	logger *log.Logger,
	serverStartArgs []string,
	serverDirectory string,
) *ProcessServer {
	s := ProcessServer{}
	s.state = serverPkg.Stopped
	s.logger = logger
	s.serverStartArgs = serverStartArgs
	s.serverDirectory = serverDirectory
	s.lines = make(chan string)
	return &s
}

// Start implements server.Server.
func (s *ProcessServer) Start() error {

	// Make the command
	// Set process group (child process dies when parent process dies)
	s.cmd = exec.Command(s.serverStartArgs[0], s.serverStartArgs[1:]...)
	s.cmd.Dir = s.serverDirectory
	s.cmd.SysProcAttr = newProcessGroup()

	err := func() error {

		var err error

		// Pipe stdin, stdout, and stderr
		s.stdin, err = s.cmd.StdinPipe()
		if err != nil {
			return err
		}
		s.stdout, err = s.cmd.StdoutPipe()
		if err != nil {
			return err
		}
		s.stderr, err = s.cmd.StderrPipe()
		if err != nil {
			return err
		}

		// Start goroutines to listen to output from stdout, stdout,
		// and watch for process exit
		s.wg.Add(2)
		go s.listenOutput(s.stdout, true)
		go s.listenOutput(s.stderr, false)
		go s.listenExit()

		// Start the command
		return s.cmd.Start()

	}()
	if err != nil {
		s.logger.Printf("error starting server: %s\n", err)
	}

	// Set state to Starting
	s.state = serverPkg.Starting

	return err

}

// Stop implements server.Server.
func (s *ProcessServer) Stop() error {

	err := func() error {

		// Check for error cases
		switch {
		case s.state == serverPkg.Stopped:
			return fmt.Errorf("tried to stop stopped server")
		case s.cmd == nil || s.cmd.Process == nil:
			return fmt.Errorf("tried to stop server with missing process")
		}

		// Execute Minecraft stop command
		_, err := s.Execute(serverPkg.StopCommand)
		if err != nil {
			return err
		}

		// Set state to Stopping
		// Wait for process exit
		s.state = serverPkg.Stopping
		<-s.exited
		return nil

	}()
	if err != nil {
		s.logger.Printf("error stopping server: %s\n", err)
	}

	return err

}

// Execute implements server.Server.
func (s *ProcessServer) Execute(command string) (string, error) {

	// Check for error case
	if s.state != serverPkg.Running {
		return "", fmt.Errorf("tried to execute on server that is not running")
	}

	// Send command to stdin
	_, err := s.stdin.Write([]byte(command + "\n"))
	if err != nil {
		return "", err
	}

	// Wait for line from stdout
	return <-s.lines, nil

}

// State implements server.Server.
func (s *ProcessServer) State() serverPkg.ServerState {
	return s.state
}

// listenOutput listens to and handles the outputs of stdout and stder.
func (s *ProcessServer) listenOutput(r io.Reader, stdout bool) {

	defer s.wg.Done()

	// Scan lines from reader
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {

		line := scanner.Text()

		// Log line
		var pid string
		if s.cmd == nil || s.cmd.Process == nil {
			pid = "unknown"
		} else {
			pid = strconv.Itoa(s.cmd.Process.Pid)
		}
		s.logger.Printf("[%s(%s)] %s", s.serverStartArgs[0], pid, line)

		// Interpret line only for stdout
		if stdout {

			// Check if startup is complete
			if s.state == serverPkg.Starting &&
				strings.Contains(line, "INFO") &&
				strings.Contains(line, "Done") {

				// Set state to Running
				s.state = serverPkg.Running

			}

			// Signal line to channel
			select {
			case s.lines <- line:
			default:
			}

		}

	}

}

// listenExit listens for the process to exit.
func (s *ProcessServer) listenExit() {

	// Wait for stdout and strerr listeners
	// Wait for process to exit
	s.wg.Wait()
	err := s.cmd.Wait()
	if err != nil {
		s.logger.Printf("server process exited with error: %s\n", err)
	} else {
		s.logger.Printf("server process exited")
	}

	// Set state to Stopped
	s.state = serverPkg.Stopped

	// Signal process exited
	select {
	case s.exited <- true:
	default:
	}

}
