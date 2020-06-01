package main

// An example SFTP server implementation using the golang SSH package.
// Serves the whole filesystem visible to the user, and has a hard-coded username and password,
// so not for real use!

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

var (
	readOnly    bool
	debugStderr bool
	debugStream io.Writer
	infoLog     *log.Logger
	errorLog    *log.Logger
)

// Based on example server code from golang.org/x/crypto/ssh and server_standalone
func main() {
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	flag.BoolVar(&readOnly, "R", false, "read-only server")
	flag.BoolVar(&debugStderr, "e", false, "debug to stderr")
	flag.Parse()

	infoLog.Printf("readonly    = %v", readOnly)
	infoLog.Printf("debugStderr = %v", debugStderr)

	debugStream = ioutil.Discard
	if debugStderr {
		debugStream = os.Stderr
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			fmt.Fprintf(debugStream, "Login: %s\n", c.User())
			if c.User() == "testuser" && string(pass) == "tiger" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		errorLog.Fatal("Failed to load private key", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		errorLog.Fatal("Failed to parse private key", err)
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		errorLog.Fatal("failed to listen for connection", err)
	}
	infoLog.Printf("Listening on %v\n", listener.Addr())
	for {
		nConn, err := listener.Accept()
		if err != nil {
			errorLog.Fatal("failed to accept incoming connection", err)
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		sconn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			errorLog.Fatal("failed to handshake", err)
		}
		infoLog.Println("login detected:", sconn.User())
		infoLog.Println("SSH server established")

		// The incoming Request channel must be serviced.
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go handleChannels(chans)
	}

}
