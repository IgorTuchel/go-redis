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
	s := store.New()

	aof, err := store.NewAOF("redis.aof")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer aof.Close()

	err = aof.Read(func(value parser.Value) {
		commandDispatch(value, s)
	})
	if err != nil {
		fmt.Println("error reading AOF: ", err.Error())
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("error accepting connection: ", err.Error())
		}
		go handleConn(conn, s, aof)
	}
}

func handleConn(conn net.Conn, s *store.Store, aof *store.AOF) {
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
		result, aofEntry := commandDispatch(value, s)
		writer.Write(result)
		for _, entry := range aofEntry {
			if err := aof.Write(entry); err != nil {
				fmt.Println("error writing to AOF: ", err.Error())
			}
		}
	}
}

func commandDispatch(value parser.Value, s *store.Store) (parser.Value, []parser.Value) {
	if len(value.Array) == 0 {
		return parser.Value{RType: parser.ERROR, Str: "ERR invalid command"}, []parser.Value{}
	}
	command := strings.ToUpper(value.Array[0].Bulk)
	args := value.Array[1:]
	switch command {
	case "SET":
		if len(args) < 2 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'set' command"}, []parser.Value{}
		}
		s.Set(args[0].Bulk, args[1].Bulk)

		aofEntry := []parser.Value{
			{
				RType: parser.ARRAY,
				Array: []parser.Value{
					{RType: parser.BULK, Bulk: "SET"},
					{RType: parser.BULK, Bulk: args[0].Bulk},
					{RType: parser.BULK, Bulk: args[1].Bulk},
				},
			},
		}

		if len(args) >= 4 {
			flag := strings.ToUpper(args[2].Bulk)
			ttl, err := strconv.Atoi(args[3].Bulk)
			if err != nil || ttl <= 0 {
				return parser.Value{RType: parser.ERROR, Str: "ERR invalid expire time in 'set' command"}, []parser.Value{}
			}
			switch flag {
			case "PX":
				ttl := time.Duration(ttl) * time.Millisecond
				ok, expiry := s.Expire(args[0].Bulk, ttl)
				if !ok {
					return parser.Value{RType: parser.ERROR, Str: "ERR failed to set expiry"}, []parser.Value{}
				}
				aofEntry = append(aofEntry, parser.Value{
					RType: parser.ARRAY,
					Array: []parser.Value{
						{RType: parser.BULK, Bulk: "EXPIREAT"},
						{RType: parser.BULK, Bulk: args[0].Bulk},
						{RType: parser.BULK, Bulk: strconv.FormatInt(expiry.UnixMilli(), 10)},
					},
				})
			case "EX":
				ttl := time.Duration(ttl) * time.Second
				ok, expiry := s.Expire(args[0].Bulk, ttl)
				if !ok {
					return parser.Value{RType: parser.ERROR, Str: "ERR failed to set expiry"}, []parser.Value{}
				}
				aofEntry = append(aofEntry, parser.Value{
					RType: parser.ARRAY,
					Array: []parser.Value{
						{RType: parser.BULK, Bulk: "EXPIREAT"},
						{RType: parser.BULK, Bulk: args[0].Bulk},
						{RType: parser.BULK, Bulk: strconv.FormatInt(expiry.Unix(), 10)},
					},
				})
			default:
				return parser.Value{RType: parser.ERROR, Str: "ERR unknown flag in 'set' command"}, []parser.Value{}
			}
		}
		return parser.Value{RType: parser.STRING, Str: "OK"}, aofEntry
	case "EXPIRE":
		if len(args) < 2 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'expire' command"}, []parser.Value{}
		}
		ttl, err := strconv.Atoi(args[1].Bulk)
		if err != nil || ttl <= 0 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid expire time in 'expire' command"}, []parser.Value{}
		}
		ok, expiry := s.Expire(args[0].Bulk, time.Duration(ttl)*time.Second)
		if !ok {
			return parser.Value{RType: parser.INTEGER, Num: 0}, []parser.Value{}
		}
		aofEntry := []parser.Value{{
			RType: parser.ARRAY,
			Array: []parser.Value{
				{RType: parser.BULK, Bulk: "EXPIREAT"},
				{RType: parser.BULK, Bulk: args[0].Bulk},
				{RType: parser.BULK, Bulk: strconv.FormatInt(expiry.Unix(), 10)},
			},
		}}
		return parser.Value{RType: parser.INTEGER, Num: 1}, aofEntry
	// TODO: need a pexpireat later
	case "EXPIREAT":
		if len(args) < 2 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'expireat' command"}, []parser.Value{}
		}
		expiry, err := strconv.ParseInt(args[1].Bulk, 10, 64)
		if err != nil {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid expire time in 'expireat' command"}, []parser.Value{}
		}
		ok := s.ExpireAt(args[0].Bulk, time.Unix(expiry, 0))
		if !ok {
			return parser.Value{RType: parser.INTEGER, Num: 0}, []parser.Value{}
		}
		aofEntry := []parser.Value{{
			RType: parser.ARRAY,
			Array: []parser.Value{
				{RType: parser.BULK, Bulk: "EXPIREAT"},
				{RType: parser.BULK, Bulk: args[0].Bulk},
				{RType: parser.BULK, Bulk: strconv.FormatInt(expiry, 10)},
			},
		}}
		return parser.Value{RType: parser.INTEGER, Num: 1}, aofEntry
	case "DEL":
		if len(args) < 1 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'del' command"}, []parser.Value{}
		}
		ok := s.Del(args[0].Bulk)
		if !ok {
			return parser.Value{RType: parser.INTEGER, Num: 0}, []parser.Value{}
		}
		aofEntry := []parser.Value{{
			RType: parser.ARRAY,
			Array: []parser.Value{
				{RType: parser.BULK, Bulk: "DEL"},
				{RType: parser.BULK, Bulk: args[0].Bulk},
			},
		}}
		return parser.Value{RType: parser.INTEGER, Num: 1}, aofEntry
	case "GET":
		if len(args) < 1 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'get' command"}, []parser.Value{}
		}
		val, ok := s.Get(args[0].Bulk)
		if !ok {
			return parser.Value{RType: parser.NULL}, []parser.Value{}
		}
		return parser.Value{RType: parser.BULK, Bulk: val}, []parser.Value{}
	case "EXISTS":
		if len(args) < 1 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'exists' command"}, []parser.Value{}
		}
		ok := s.Exists(args[0].Bulk)
		if !ok {
			return parser.Value{RType: parser.INTEGER, Num: 0}, []parser.Value{}
		}
		return parser.Value{RType: parser.INTEGER, Num: 1}, []parser.Value{}
	case "TTL":
		if len(args) < 1 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'ttl' command"}, []parser.Value{}
		}
		ttl := s.Ttl(args[0].Bulk)
		return parser.Value{RType: parser.INTEGER, Num: ttl}, []parser.Value{}
	case "PING":
		return parser.Value{RType: parser.STRING, Str: "PONG"}, []parser.Value{}
	case "ECHO":
		if len(args) < 1 {
			return parser.Value{RType: parser.ERROR, Str: "ERR invalid number of arguments for 'echo' command"}, []parser.Value{}
		}
		return parser.Value{RType: parser.BULK, Bulk: args[0].Bulk}, []parser.Value{}
	default:
		return parser.Value{RType: parser.ERROR, Str: "ERR unknown command"}, []parser.Value{}
	}
}
