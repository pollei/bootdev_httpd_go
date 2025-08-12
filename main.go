package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	retChan := make(chan string)
	go func() {
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
		f.Close()
		close(retChan)
	}()
	return retChan
}

func main() {
	messFile, err := os.Open("public_html/messages.txt")
	if err != nil {
		fmt.Println("shot me now")
		return
	}
	linesChan := getLinesChannel(messFile)
	for line := range linesChan {
		fmt.Printf("read: %s\n", line)
	}

}
func origMain() {

	messFile, err := os.Open("public_html/messages.txt")
	if err != nil {
		fmt.Println("shot me now")
		return
	}
	buf := make([]byte, 8)
	var prevStr string
	for {
		n, err := messFile.Read(buf)
		curStr := string(buf[0:n])
		lines := strings.Split(curStr, "\n")
		lines[0] = prevStr + lines[0]
		for i := 0; i < len(lines)-1; i++ {
			fmt.Printf("read: %s\n", lines[i])
		}
		prevStr = lines[len(lines)-1]
		if err == io.EOF {
			break
		}
		//fmt.Printf("read: %s\n", string(buf[0:n]))
	}
	if len(prevStr) > 0 {
		fmt.Printf("read: %s\n", prevStr)
	}

}
