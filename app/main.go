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
		// fmt.Println(cmd, arg)
		if cmd == "echo" {
			msg := fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
			_, err := conn.Write([]byte(msg))
			if err != nil {
				fmt.Println("Not able to echo the msg")
			}
			return
		}

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			return
		}
	}
}

func respToString(resp string) (string, string) {
	// *2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n
	// 4\r\nECHO\r\n
	indi := strings.Split(resp, "$")

	var words []string
	// if indi[0][0] == '*' {
	// 	n := int(indi[0][1] - '0')
	// }
	indi = indi[1:]

	for _, part := range indi {
		var charBuf []byte
		for i := 3; i < len(part)-2; i++ {
			charBuf = append(charBuf, part[i])
		}
		words = append(words, string(charBuf))
	}

	return strings.ToLower(words[0]), words[1]
}

// func stringToResp(msg string) string {
//
// }
