package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/igortuchel/go-redis/internal/parser"
	"github.com/igortuchel/go-redis/internal/store"
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
	store := store.New()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error accepting connection: ", err.Error())
		}
		go handleConn(conn, store)
	}
}

func handleConn(conn net.Conn, s *store.Store) {
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

		// TODO: Refactor into a command dispatcher
		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		switch command {
		case "PING":
			writer.Write(parser.Value{RType: parser.STRING, Str: "PONG"})
		case "SET":
			s.Set(args[0].Bulk, args[1].Bulk)
			if len(args) >= 4 {
				flag := strings.ToUpper(args[2].Bulk)
				ttl, err := strconv.Atoi(args[3].Bulk)
				if err != nil || ttl <= 0 {
					writer.Write(parser.Value{RType: parser.ERROR, Str: "ERR invalid expire time in 'set' command"})
					break
				}
				switch flag {
				case "PX":
					s.Expire(args[0].Bulk, time.Duration(ttl)*time.Millisecond)
				case "EX":
					s.Expire(args[0].Bulk, time.Duration(ttl)*time.Second)
				default:
					writer.Write(parser.Value{RType: parser.ERROR, Str: "ERR syntax error"})
					break
				}
			}
			writer.Write(parser.Value{RType: parser.STRING, Str: "OK"})
		case "GET":
			val, ok := s.Get(args[0].Bulk)
			if ok {
				writer.Write(parser.Value{RType: parser.BULK, Bulk: val})
			} else {
				writer.Write(parser.Value{RType: parser.NULL})
			}
		case "EXPIRE":
			ttl, err := strconv.Atoi(args[1].Bulk)
			if err != nil || ttl <= 0 {
				writer.Write(parser.Value{RType: parser.INTEGER, Num: 0})
				break
			}
			ok := s.Expire(args[0].Bulk, time.Duration(ttl)*time.Second)
			if ok {
				writer.Write(parser.Value{RType: parser.INTEGER, Num: 1})
			} else {
				writer.Write(parser.Value{RType: parser.INTEGER, Num: 0})
			}
		case "DEL":
			ok := s.Del(args[0].Bulk)
			if ok {
				writer.Write(parser.Value{RType: parser.INTEGER, Num: 1})
			} else {
				writer.Write(parser.Value{RType: parser.INTEGER, Num: 0})
			}
		case "EXISTS":
			exists := s.Exists(args[0].Bulk)
			if exists {
				writer.Write(parser.Value{RType: parser.INTEGER, Num: 1})
			} else {
				writer.Write(parser.Value{RType: parser.INTEGER, Num: 0})

			}
		case "ECHO":
			writer.Write(args[0])
		case "TTL":
			ttl := s.Ttl(args[0].Bulk)
			writer.Write(parser.Value{RType: parser.INTEGER, Num: ttl})
		default:
			writer.Write(parser.Value{RType: parser.ERROR, Str: "ERR unknown command"})
		}

	}
}
