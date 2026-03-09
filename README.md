# SFTP Server & Key Generator

A lightweight, configurable SFTP server written in Go, designed for testing, internal development, and validating SFTP client implementations. This project also includes a utility for generating RSA SSH key pairs.

<<<<<<< HEAD
**Note:** This server is intended for development and testing purposes. While it supports configuration via flags, please review security requirements before using it in a production environment.
=======
It is a lightweight SFTP server for internal use. I created it to test and validate some other developments. The code is based on the code found in the SFTP library documentation. I modified it so that it doesn't stop after the first connection.  I tested it on Windows and Linux, and on Mac OS. I recommend not to use it in production because the user and password are hardcoded in the code. 
>>>>>>> 5c9b8e1252df6b06c98ff76002a211a2f1b87aab

## Features

* **Configurable**: Customize listening address, port, credentials, and root directory via command-line flags.
* **Cross-Platform**: Works on Windows, Linux, and macOS.
* **Concurrent Connections**: Handles multiple client connections simultaneously.
* **Read-Only Mode**: Optional flag to restrict clients to read operations.
* **Key Generator**: Includes a tool to generate RSA private/public key pairs.

## Building

To compile the binaries, run the following commands from the project root:

```bash
    go build ./cmd/sftpserver/
    go build ./cmd/keygen/
```

## Usage

### 1. Generate SSH Keys

The SFTP server requires an SSH private key to start. You can use the included `keygen` tool to generate one.

```bash
# Generate a key pair (defaults to ./id_rsa and ./id_rsa.pub)
./keygen

# Or specify output files and bit size
./keygen -o my_key -pub my_key.pub -b 2048
```

**Keygen Flags:**

* `-o`: Path to save the private key (default "id_rsa")
* `-pub`: Path to save the public key (default "id_rsa.pub")
* `-b`: Number of bits for the key (default 4096)

### 2. Run the SFTP Server

Once you have a private key (default expected is `id_rsa` in the current directory), you can start the server.

```bash
# Run with a randomly generated password (user: oracle, port: 2022)
# The server will print the generated password to the console on startup.
./sftpserver

# Run with custom configuration
./sftpserver -port 2222 -user myuser -pass mypassword -root ./files -R
```

**Server Flags:**

* `-host`: Host to listen on (default "0.0.0.0")
* `-port`: Port to listen on (default "2022")
* `-user`: Username for authentication (default "oracle")
* `-pass`: Password for authentication. If not provided, a random password will be generated and printed on startup.
* `-root`: SFTP root directory (default "./in"). The server will create this directory if it doesn't exist.
* `-keyfile`: Path to SSH private key (default "id_rsa")
* `-R`: Enable read-only server mode (default false)
* `-e`: Enable debug logging to stderr (default false)

### 3. Connect with a Client

You can connect using any standard SFTP client.

```bash
sftp -P 2022 oracle@localhost
```

## Acknowledgements

This project is based on the example server code from the pkg/sftp library and the Go SSH package.
