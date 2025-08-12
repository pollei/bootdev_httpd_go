package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	retChan := make(chan string)
	go func() {
		defer f.Close()
		defer close(retChan)
		buf := make([]byte, 8)
		var prevStr string
		for {
			n, err := f.Read(buf)
			curStr := string(buf[0:n])
			lines := strings.Split(curStr, "\n")
			lines[0] = prevStr + lines[0]
			for i := 0; i < len(lines)-1; i++ {
				retChan <- lines[i]
			}
			prevStr = lines[len(lines)-1]
			if err == io.EOF {
				break
			}
		}
		if len(prevStr) > 0 {
			retChan <- prevStr
		}
	}()
	return retChan
}

func main() {
	listenPort, err := net.Listen("tcp4", "127.0.0.1:42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listenPort.Close()
	for {
		conn, err := listenPort.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			defer c.Close()
			linesChan := getLinesChannel(c)
			for line := range linesChan {
				fmt.Printf("%s\n", line)
			}
		}(conn)
	}

}
