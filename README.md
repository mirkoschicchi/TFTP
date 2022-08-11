# TFTP client and server
## Introduction
This is my personal implementation in Golang of the TFTP protocol based on [https://www.ietf.org/rfc/rfc1350.txt](https://www.ietf.org/rfc/rfc1350.txt) .

Since this is just a project done for fun, many feature are not fully implemented.

## Build
In order to build the project run the command
```bash
go build -o tftp cmd/main.go
```

## Launch the server
Navigate to the directory that will become the base directory for the server and use the command
```bash
sudo ./tftp -server
```

The server will listen on address **127.0.0.1:69** .

## Launch the client
The client can either write or request a file from the server.

### Write a file to server
In order to write a file to the server use the following command:

```bash
./tftp -remote="127.0.0.1:69" -client -read <path_to_file>
```
The command will save the retrieved file to the current directory from which the client has been launched.

### Read a file from the server
In order to read a file from the server use the following command:

```bash
./tftp -remote="127.0.0.1:69" -client -write <path_to_file>
```
The command will retrieve a file from the client side and store it into the server main directory.
