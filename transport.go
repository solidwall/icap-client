package icapclient

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"
	"time"
)

const (
	MaxReadSocketLength = 1096
)

// transport represents the transport layer data
type transport struct {
	network      string
	addr         string
	timeout      time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	sckt         net.Conn
}

// dial fires up a tcp socket
func (t *transport) dial() error {
	sckt, err := net.DialTimeout(t.network, t.addr, t.timeout)

	if err != nil {
		return err
	}

	if err := sckt.SetReadDeadline(time.Now().UTC().Add(t.readTimeout)); err != nil {
		return err
	}

	if err := sckt.SetWriteDeadline(time.Now().UTC().Add(t.writeTimeout)); err != nil {
		return err
	}

	t.sckt = sckt

	return nil
}

// dialWithContext fires up a tcp socket
func (t *transport) dialWithContext(ctx context.Context) error {
	sckt, err := (&net.Dialer{
		Timeout: t.timeout,
	}).DialContext(ctx, t.network, t.addr)

	if err != nil {
		return err
	}

	if err := sckt.SetReadDeadline(time.Now().UTC().Add(t.readTimeout)); err != nil {
		return err
	}

	if err := sckt.SetWriteDeadline(time.Now().UTC().Add(t.writeTimeout)); err != nil {
		return err
	}

	t.sckt = sckt

	return nil
}

// Write writes data to the server
func (t *transport) write(data []byte) (int, error) {
	logDebug("Dumping the message being sent to the server...")
	dumpDebug(string(data))
	return t.sckt.Write(data)
}

type readingState int64 //states for reading "Encapsulated" header value

const (
	Identifier readingState = iota
	Number
)

func isLetter(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isNumber(c rune) bool {
	return ('0' <= c && c <= '9')
}

func findLastSectionStart(str []byte) int {
	encapsulatedPos := strings.Index(string(str), EncapsulatedHeader)
	maxNum := 0
	s := bytes.Runes(str)
	if encapsulatedPos < 0 {
		return 0
	}

	st := Identifier
	num := 0
	for i := encapsulatedPos + len(EncapsulatedHeader) + 1; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			continue
		}
		if s[i] == '\r' { //ending of Encapsulated header line
			break
		}
		switch st {
		case Identifier:
			if s[i] == '=' {
				st = Number
				num = 0
			} else if !(isLetter(s[i]) || s[i] == '-') {
				logDebug("identifier followed by ", s[i])
				return 0
			}
		case Number:
			if isNumber(s[i]) {
				num = num*10 + int(s[i]-'0')
			} else if s[i] == ',' {
				st = Identifier
				if num > maxNum {
					maxNum = num
				}
			} else {
				logDebug("number followed by ", s[i])
				return 0
			}
		default:
			logDebug("undefined reading state")
			return 0
		}
	}
	return maxNum
}

// Read reads data from server
func (t *transport) read() (string, error) {

	data := make([]byte, 0)

	logDebug("Dumping messages received from the server...")

	for {
		tmp := make([]byte, MaxReadSocketLength)

		n, err := t.sckt.Read(tmp)

		if err != nil {
			if err == io.EOF {
				logDebug("End of file detected from EOF error")
				break
			}
			return "", err
		}

		if n == 0 {
			logDebug("End of file detected by 0 bytes")
			break
		}

		data = append(data, tmp[:n]...)
		expectedLength := findLastSectionStart(tmp) // this is the beginning of last section in file
		// sections are separated with `\r\n\r\n` (DoubleCRLF)
		if len(data) < expectedLength {
			dumpDebug(string(tmp))
			continue
		}

		if strings.HasSuffix(string(data), DoubleCRLF) {
			logDebug("End of the file detected by Double CRLF indicator")
			break
		}

	}
	return string(data), nil
}

// close closes the tcp connection
func (t *transport) close() error {
	return t.sckt.Close()
}
