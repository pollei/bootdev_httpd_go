package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestScanToken(t *testing.T) {
	adv, tok, err := scanToken([]byte("GET "), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 4)
	assert.Equal(t, tok, []byte("GET"))

	adv, tok, err = scanToken([]byte("GET\n"), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 3)
	assert.Equal(t, tok, []byte("GET"))

	adv, _, err = scanToken([]byte("GET"), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 0)

	adv, tok, err = scanToken([]byte("GET"), true)
	require.NoError(t, err)
	assert.Equal(t, adv, 3)
	assert.Equal(t, tok, []byte("GET"))

}

func TestScanAsciiPrintable(t *testing.T) {
	adv, tok, err := scanAsciiPrintable([]byte("/ "), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 2)
	assert.Equal(t, tok, []byte("/"))

	adv, tok, err = scanAsciiPrintable([]byte("/\n"), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 1)
	assert.Equal(t, tok, []byte("/"))

	adv, _, err = scanAsciiPrintable([]byte("/coffee"), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 0)
	//assert.Equal(t, tok, []byte("/coffee"))

	adv, tok, err = scanAsciiPrintable([]byte("/coffee"), true)
	require.NoError(t, err)
	assert.Equal(t, adv, 7)
	assert.Equal(t, tok, []byte("/coffee"))

}

func TestScanCrLf(t *testing.T) {
	adv, _, err := scanCrLf([]byte("\r\n"), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 2)
	adv, _, err = scanCrLf([]byte("\n"), false)
	require.NoError(t, err)
	assert.Equal(t, adv, 1)
}

func TestRequestLineParse(t *testing.T) {
	assert.Equal(t, "TheTestagen", "TheTestagen")
	// Test: Good GET Request line
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}
