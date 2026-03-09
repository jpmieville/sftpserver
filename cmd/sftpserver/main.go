package main

// An example SFTP server implementation using the golang SSH package.
// Serves files from a defined root directory.

import (
	"crypto/subtle"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

var (
	readOnly    bool
	debugStderr bool
	sftpRoot    string
	host        string
	port        string
	user        string
	pass        string
	keyFile     string
	debugStream io.Writer
	infoLog     *log.Logger
	errorLog    *log.Logger
)

// Based on example server code from golang.org/x/crypto/ssh and server_standalone
func main() {
	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	flag.BoolVar(&readOnly, "R", false, "Read-only server")
	flag.BoolVar(&debugStderr, "e", false, "Enable debug logging to stderr")
	flag.StringVar(&sftpRoot, "root", "./in", "SFTP root directory")
	flag.StringVar(&host, "host", "0.0.0.0", "Host to listen on")
	flag.StringVar(&port, "port", "2022", "Port to listen on")
	flag.StringVar(&user, "user", "oracle", "Username for authentication")
	flag.StringVar(&pass, "pass", "jpmorgan4oracle", "Password for authentication")
	flag.StringVar(&keyFile, "keyfile", "id_rsa", "Path to SSH private key")
	flag.Parse()

	// Setup SFTP root directory. This is a process-wide operation.
	if _, err := os.Stat(sftpRoot); os.IsNotExist(err) {
		if err := os.MkdirAll(sftpRoot, 0755); err != nil {
			errorLog.Fatalf("Failed to create root directory %s: %v", sftpRoot, err)
		}
		infoLog.Printf("Created directory '%s'", sftpRoot)
	}
	if err := os.Chdir(sftpRoot); err != nil {
		errorLog.Fatalf("Failed to change directory to %s: %v", sftpRoot, err)
	}
	currentDir, _ := os.Getwd()
	infoLog.Printf("SFTP root is '%s'", currentDir)

	infoLog.Printf("Read-only server: %v", readOnly)
	infoLog.Printf("Debug logging: %v", debugStderr)

	debugStream = io.Discard
	if debugStderr {
		debugStream = os.Stderr
	}

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, passwordBytes []byte) (*ssh.Permissions, error) {
			fmt.Fprintf(debugStream, "Login attempt for user %s\n", c.User())
			// Use constant-time comparison to prevent timing attacks.
			userMatch := subtle.ConstantTimeCompare([]byte(c.User()), []byte(user)) == 1
			passMatch := subtle.ConstantTimeCompare(passwordBytes, []byte(pass)) == 1

			if userMatch && passMatch {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := os.ReadFile(keyFile)
	if err != nil {
		errorLog.Fatalf("Failed to load private key from %s: %v", keyFile, err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		errorLog.Fatal("Failed to parse private key", err)
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	addr := net.JoinHostPort(host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		errorLog.Fatalf("Failed to listen for connection on %s: %v", addr, err)
	}
	infoLog.Printf("Listening on %s", addr)
	for {
		nConn, err := listener.Accept()
		if err != nil {
			errorLog.Printf("Failed to accept incoming connection: %v", err)
			continue // Continue accepting connections
		}

		// Before use, a handshake must be performed on the incoming
		// net.Conn.
		sconn, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			errorLog.Print("failed to handshake", err)
			continue
		}
		infoLog.Printf("Login successful for user: %s from %s", sconn.User(), sconn.RemoteAddr())
		infoLog.Println("SSH server established")

		// The incoming Request channel must be serviced.
		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go handleChannels(chans)
	}

}
