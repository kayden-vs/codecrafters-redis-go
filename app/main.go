package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	// hold the main thread until os signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {

		for {
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting connection: ", err.Error())
				os.Exit(1)
			}

			go handleClient(conn)
		}
	}()

	<-ctx.Done()
	l.Close()
}

func handleClient(connection net.Conn) {
	defer connection.Close()

	db := make(map[string]string)
	var buffer [128]byte
	for {
		n, err := connection.Read(buffer[:])
		if err == io.EOF {
			break
		}

		command := decodeBulkStringArray(buffer[:n])

		if command == nil {
			continue
		}

		switch strings.ToLower(command[0]) {
		case "ping":
			connection.Write(encodeSimpleString("PONG"))
		case "echo":
			connection.Write(encodeBulkString(command[1]))
		case "set":
			db[command[1]] = command[2]
			connection.Write(encodeSimpleString("OK"))
		case "get":
			connection.Write(encodeBulkString(db[command[1]]))
		}
	}
}

func decodeBulkStringArray(data []byte) []string {
	if len(data) == 0 {
		return nil
	}

	if data[0] != '*' {
		log.Fatal("Failed to decode bulk string array")
	}

	var lines []string

	_, n := getLine(data)

	for n < len(data) {
		line, length := decodeBulkString(data[n:])
		lines = append(lines, line)
		n += length
	}

	return lines
}

func decodeBulkString(data []byte) (string, int) {
	if data[0] != '$' {
		log.Fatal("Failed to decode bulk string")
	}

	_, start := getLine(data)
	line, length := getLine(data[start:])

	return string(line), start + length
}

func getLine(data []byte) ([]byte, int) {
	i := 0

	for data[i] != '\r' || data[i+1] != '\n' {
		i++
	}

	return data[:i], i + 2
}

func encodeSimpleString(str string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", str))
}

func encodeBulkString(str string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
}
