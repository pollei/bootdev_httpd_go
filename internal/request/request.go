package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var _ = fmt.Printf // For debugging;

const (
	RequestProgressNone = iota
	RequestProgressMethod
	RequestProgressTarget
	RequestProgressProtocol
	RequestProgressDone
)

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestError struct {
	Code string
	Err  error
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// https://go.dev/blog/go1.13-errors
// https://pkg.go.dev/errors
// 505 Version not supported
// 501 method not implemented
// 400 bad request
// 405 method not allowed
// 414 uri too long
func (e *RequestError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *RequestError) Unwrap() error { return e.Err }

func isTokenByte(b byte) bool {
	if b >= '^' && b <= 'z' {
		return true
	}
	if b >= 'A' && b <= 'Z' {
		return true
	}
	if b >= '0' && b <= '9' {
		return true
	}
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*',
		'+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}

func isWhitespaceByte(b byte) bool {
	switch b {
	case ' ', '\r', '\f', '\t':
		return true
	default:
		return false
	}
}

// func(data []byte, atEOF bool) (advance int, token []byte, err error)

func scanToken(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advIndx, dataSz := 0, len(data)
	for ; advIndx < dataSz; advIndx++ {
		currByte := data[advIndx]
		if isTokenByte(currByte) {
			continue
		}
		if isWhitespaceByte(currByte) {
			return advIndx, data[0:advIndx], nil
		}
		if currByte == '\n' {
			return advIndx - 1, data[0:advIndx], nil
		}
		return 0, nil, errors.New("invalid token")
	}
	return 0, nil, nil
}

// not white-space visible ascii chars
func scanAsciiPrintable(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advIndx, dataSz := 0, len(data)
	for ; advIndx < dataSz; advIndx++ {
		currByte := data[advIndx]
		if currByte >= '!' && currByte <= '~' {
			continue
		}
		if isWhitespaceByte(currByte) {
			return advIndx, data[0:advIndx], nil
		}
		if currByte == '\n' {
			return advIndx - 1, data[0:advIndx], nil
		}
		return 0, nil, errors.New("invalid printable")
	}
	return 0, nil, nil
}

func scanCrLf(data []byte, atEOF bool) (advance int, token []byte, err error) {
	dataSz := len(data)
	if dataSz >= 1 && data[0] == '\n' {
		return 1, nil, nil
	}
	if dataSz >= 2 && data[0] == '\r' && data[1] == '\n' {
		return 2, nil, nil
	}
	if (dataSz >= 1 && data[0] != '\r') || dataSz >= 2 {
		return 0, nil, errors.New("invalid EOL")
	}
	return 0, nil, nil
}

func (r *Request) parse(data []byte) (int, error) {
	return 0, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	rawBuf, err := io.ReadAll(reader)
	if err != nil {
		return nil, &RequestError{Code: "500", Err: err}
	}
	if rawBuf == nil {
		return nil, &RequestError{Code: "400", Err: errors.New("empty request")}
	}
	ret := Request{}
	if len(rawBuf) < 12 {
		return nil, &RequestError{Code: "400", Err: errors.New("too short request")}
	}
	methodEndIndx := 0
	rawBufSz := len(rawBuf)
	maxMethSz := max(1, min(72, rawBufSz-8))
	for ; methodEndIndx < maxMethSz; methodEndIndx++ {
		if isTokenByte(rawBuf[methodEndIndx]) {
			continue
		}
		if rawBuf[methodEndIndx] == ' ' {
			if methodEndIndx <= 0 {
				return nil, &RequestError{Code: "400", Err: errors.New("no method in request")}
			}
			ret.RequestLine.Method = string(rawBuf[0:methodEndIndx])
			break
		}
		return nil, &RequestError{Code: "400", Err: errors.New("bogus request")}
	}
	//fmt.Println("after method")
	if rawBuf[methodEndIndx] != ' ' {
		return nil, &RequestError{Code: "400", Err: errors.New("bogus request")}
	}
	targetEndIndx := methodEndIndx + 1
	maxTargetSz := min(8192, rawBufSz-6)
	for ; targetEndIndx < maxTargetSz; targetEndIndx++ {
		currByte := rawBuf[targetEndIndx]
		if currByte >= '!' && currByte <= '~' {
			continue
		}
		if currByte == ' ' {
			if targetEndIndx <= methodEndIndx+1 {
				return nil, &RequestError{Code: "400", Err: errors.New("no target in request")}
			}
			ret.RequestLine.RequestTarget = string(rawBuf[methodEndIndx+1 : targetEndIndx])
			break
		}
		return nil, &RequestError{Code: "400", Err: errors.New("bogus request")}
	}
	//fmt.Println("after target")
	if rawBuf[targetEndIndx] != ' ' || targetEndIndx+9 > rawBufSz {
		return nil, &RequestError{Code: "400", Err: errors.New("bogus request")}
	}
	if !bytes.Equal([]byte("HTTP/"), rawBuf[targetEndIndx+1:targetEndIndx+6]) {
		return nil, &RequestError{Code: "400", Err: errors.New("bogus protocol in request")}

	}
	majorVer, minorVer := rawBuf[targetEndIndx+6], rawBuf[targetEndIndx+8]
	if majorVer < '0' || majorVer > '9' ||
		minorVer < '0' || minorVer > '9' || rawBuf[targetEndIndx+7] != '.' {
		return nil, &RequestError{Code: "400", Err: errors.New("bogus protocol version in request")}
	}
	ret.RequestLine.HttpVersion = string(rawBuf[targetEndIndx+6 : targetEndIndx+9])
	return &ret, nil
}
