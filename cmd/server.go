package main

import (
	"fmt"
	"io"
	"net"

	"github.com/igortuchel/go-redis/internal/parser"
)

func main() {
	fmt.Println("Starting GoRedis")

	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer l.Close()

	fmt.Println("Server running on port :6379")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error accepting connection: ", err.Error())
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	writer := parser.NewWriter(conn)
	reader := parser.NewResp(conn)

	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("error reading: ", err.Error())
			return
		}

		fmt.Println("recieved: ", value)

		writer.Write(parser.Value{RType: parser.STRING, Str: "OK"})
	}
}
