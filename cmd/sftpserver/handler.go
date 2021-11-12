package main

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	// Service the incoming Channel channel.
	// Channels have a type, depending on the application level
	// protocol intended. In the case of an SFTP session, this is "subsystem"
	// with a payload string of "<length=4>sftp"
	fmt.Fprintf(debugStream, "Incoming channel: %s\n", newChannel.ChannelType())
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		fmt.Fprintf(debugStream, "Unknown channel type: %s\n", newChannel.ChannelType())
		return
	}
	channel, requests, err := newChannel.Accept()
	if err != nil {
		infoLog.Print("could not accept channel.", err)
		return
	}
	fmt.Fprintf(debugStream, "Channel accepted\n")

	// Sessions have out-of-band requests such as "shell",
	// "pty-req" and "env".  Here we handle only the
	// "subsystem" request.
	go func(in <-chan *ssh.Request) {
		for req := range in {
			fmt.Fprintf(debugStream, "Request: %v\n", req.Type)
			ok := false
			switch req.Type {
			case "subsystem":
				fmt.Fprintf(debugStream, "Subsystem: %s\n", req.Payload[4:])
				if string(req.Payload[4:]) == "sftp" {
					ok = true
				}
			}
			fmt.Fprintf(debugStream, " - accepted: %v\n", ok)
			req.Reply(ok, nil)
		}
	}(requests)

	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(debugStream),
	}

	if readOnly {
		serverOptions = append(serverOptions, sftp.ReadOnly())
		fmt.Fprintf(debugStream, "Read-only server\n")
	} else {
		fmt.Fprintf(debugStream, "Read write server\n")
	}

	serverOptions = append(serverOptions, setDir())

	server, err := sftp.NewServer(
		channel,
		serverOptions...,
	)
	if err != nil {
		infoLog.Print(err)
		return
	}
	defer server.Close()
	if err := server.Serve(); err == io.EOF {
		server.Close()
		infoLog.Print("sftp client exited session.")
	} else if err != nil {
		infoLog.Print("sftp server completed with error:", err)
		return
	}
}

func setDir() sftp.ServerOption {
	err := os.Chdir("./in")
	if err != nil {
		err2 := os.Mkdir("./in", 0755)
		if err2 != nil {
			errorLog.Fatalf("sftpserver could not create the './in' directory: %v : %v", err, err2)
		}
		infoLog.Print("directory './in' created")
	}
	infoLog.Print("chdir to in")
	currentDir, err2 := os.Getwd()
	if err2 != nil {
		errorLog.Printf("sftpserver could not list the current working directory: %v", err2)
	}
	infoLog.Printf("current working directory '%s'", currentDir)
	return func(s *sftp.Server) error {
		return nil
	}
}
