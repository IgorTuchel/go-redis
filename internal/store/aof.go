package store

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"

	"github.com/igortuchel/go-redis/internal/parser"
)

type AOF struct {
	mu   sync.Mutex
	file *os.File
	buf  *bufio.Writer
	done chan struct{}
}

// TODO: Read function
// Read from a file
// Update the instance
func (a *AOF) Read(fn func(parser.Value)) error {
	a.mu.Lock()
	a.buf.Flush()
	_, err := a.file.Seek(0, io.SeekStart)
	a.mu.Unlock()

	if err != nil {
		return err
	}

	reader := parser.NewResp(a.file)
	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		fn(value)
	}
	return nil
}

func NewAOF(path string) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	aof := &AOF{
		file: f,
		buf:  bufio.NewWriter(f),
		done: make(chan struct{}),
	}
	go aof.fsyncEverySecond()
	return aof, nil
}

// TODO: flush buffer and sync to disk
func (a *AOF) fsyncEverySecond() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.mu.Lock()
			a.buf.Flush()
			a.file.Sync()
			a.mu.Unlock()
		case <-a.done:
			return
		}
	}
}

func (a *AOF) Close() error {
	close(a.done)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.buf.Flush()
	a.file.Sync()
	return a.file.Close()
}

// TODO: Write function
func (a *AOF) Write(value parser.Value) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	v, err := value.Marshal()
	if err != nil {
		return err
	}
	_, err = a.buf.Write(v)
	if err != nil {
		return err
	}
	return nil
}

