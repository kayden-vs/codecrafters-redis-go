package main

import (
	"context"
	"fmt"
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

func handleClient(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Connection closed")
			return
		}

		cmd, arg := respToString(string(buf[:n]))
		fmt.Println(cmd, arg)

		if cmd == "ping" {
			_, err = conn.Write([]byte("+PONG\r\n"))
			if err != nil {
				fmt.Println("Error writing PONG")
			}
			return
		}

		if cmd == "echo" {
			msg := fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
			_, err := conn.Write([]byte(msg))
			if err != nil {
				fmt.Println("Not able to echo the msg")
			}
			return
		}

	}
}

func respToString(resp string) (string, string) {
	if len(resp) == 0 {
		return "", ""
	}

	// Single bulk string command: $4\r\nPING\r\n
	if resp[0] == '$' {
		lines := strings.Split(resp, "\r\n")
		if len(lines) >= 2 {
			return strings.ToLower(lines[1]), ""
		}
		return "", ""
	}

	// Standard RESP array: *1\r\n$4\r\nPING\r\n or *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
	if resp[0] == '*' {
		indi := strings.Split(resp, "$")[1:]
		var words []string

		for _, part := range indi {
			lines := strings.Split(part, "\r\n")
			if len(lines) >= 2 {
				words = append(words, lines[1])
			}
		}

		if len(words) == 0 {
			return "", ""
		}
		if len(words) == 1 {
			return strings.ToLower(words[0]), ""
		}
		return strings.ToLower(words[0]), strings.Join(words[1:], " ")
	}

	// Inline command fallback (e.g., "PING\r\n")
	parts := strings.Fields(resp)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return strings.ToLower(parts[0]), ""
	}
	return strings.ToLower(parts[0]), strings.Join(parts[1:], " ")
}

// func stringToResp(msg string) string {
//
// }
