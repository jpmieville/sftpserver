# SFTP server


It is a lightweight SFTP server for internal use. I created it to test and validate some other developments. The code is based on the code found in the SFTP library documentation. I modified it so that it doesn't stop after the first connection.  I tested it on Windows and Linux, and it should also work on Mac OS. I recommend not to use it in production because the user and password are hardcoded in the code. 

I created also a program to create RSA key. To be used on windows for example. 

All the code is in the folder cmd. 

To compile the code run the following command.

    go build ./cmd/sftpserver/
    go build ./cmd/keygen/
