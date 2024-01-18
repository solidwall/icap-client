package icapclient

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const ResponseOptions = "ICAP/1.0 200 OK\r\n" +
	"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
	"Methods: RESPMOD\r\n" +
	"Service: FOO Tech Server 1.0\r\n" +
	"ISTag: \"W3E4R7U9-L2E4-2\"\r\n" +
	"Encapsulated: null-body=0\r\n" +
	"Max-Connections: 1000\r\n" +
	"Options-TTL: 7200\r\n" +
	"Allow: 204\r\n" +
	"Preview: 2048\r\n" +
	"Transfer-Complete: asp, bat, exe, com\r\n" +
	"Transfer-Ignore: html\r\n" +
	"Transfer-Preview: *\r\n\r\n"

const ResponseReqmod1 = "ICAP/1.0 200 OK\r\n" +
	"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
	"Server: ICAP-Server-Software/1.0\r\n" +
	"Connection: close\r\n" +
	"ISTag: \"W3E4R7U9-L2E4-2\"\r\n" +
	"Encapsulated: req-hdr=0, null-body=231\r\n" +
	"\r\n" +
	"GET /modified-path HTTP/1.1\r\n" +
	"Host: www.origin-server.com\r\n" +
	"Via: 1.0 icap-server.net (ICAP Example ReqMod Service 1.1)\r\n" +
	"Accept: text/html, text/plain, image/gif\r\n" +
	"Accept-Encoding: gzip, compress\r\n" +
	"If-None-Match: \"xyzzy\", \"r2d2xxxx\"\r\n\r\n"

const ResponseReqmod2 = "ICAP/1.0 200 OK\r\n" +
	"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
	"Server: ICAP-Server-Software/1.0\r\n" +
	"Connection: close\r\n" +
	"ISTag: \"W3E4R7U9-L2E4-2\"" +
	"Encapsulated: req-hdr=0, req-body=244\r\n" +
	"\r\n" +
	"POST /origin-resource/form.pl HTTP/1.1\r\n" +
	"Host: www.origin-server.com\r\n" +
	"Via: 1.0 icap-server.net (ICAP Example ReqMod Service 1.1)\r\n" +
	"Accept: text/html, text/plain, image/gif\r\n" +
	"Accept-Encoding: gzip, compress\r\n" +
	"Pragma: no-cache\r\n" +
	"Content-Length: 45\r\n" +
	"\r\n" +
	"2d\r\n" +
	"I am posting this information.  ICAP powered!\r\n" +
	"0\r\n\r\n"

const ResponseReqmod3 = "ICAP/1.0 200 OK\r\r" +
	"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
	"Server: ICAP-Server-Software/1.0\r\n" +
	"Connection: close\r\n" +
	"ISTag: \"W3E4R7U9-L2E4-2\"\r\n" +
	"Encapsulated: res-hdr=0, res-body=213\r\n" +
	"\r\n" +
	"HTTP/1.1 403 Forbidden\r\n" +
	"Date: Wed, 08 Nov 2000 16:02:10 GMT\r\n" +
	"Server: Apache/1.3.12 (Unix)\r\n" +
	"Last-Modified: Thu, 02 Nov 2000 13:51:37 GMT\r\n" +
	"ETag: \"63600-1989-3a017169\"\r\n" +
	"Content-Length: 58\r\n" +
	"Content-Type: text/html\r\n" +
	"\r\n" +
	"3a\r\n" +
	"Sorry, you are not allowed to access that naughty content.\r\n" +
	"0\r\n\r\n"

const ResponseRespmodPart1 = "ICAP/1.0 200 OK\r\n" +
	"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
	"Server: ICAP-Server-Software/1.0\r\n" +
	"Connection: close\r\n" +
	"ISTag: \"W3E4R7U9-L2E4-2\"\r\n" +
	"Encapsulated: res-hdr=0, res-body=222\r\n" +
	"\r\n"

const ResponseRespmodPart2 =
	"HTTP/1.1 200 OK\r\n" +
	"Date: Mon, 10 Jan 2000  09:55:21 GMT\r\n" +
	"Via: 1.0 icap.example.org (ICAP Example RespMod Service 1.1)\r\n" +
	"Server: Apache/1.3.6 (Unix)\r\n" +
	"ETag: \"63840-1ab7-378d415b\"\r\n" +
	"Content-Type: text/html\r\n" +
	"Content-Length: 92\r\n" +
	"\r\n" +
	"5c\r\n" +
	"This is data that was returned by an origin server, but with\r\n" +
	"value added by an ICAP server.\r\n" +
	"0\r\n\r\n"

const NotValidResponse = "Some strange response..."

var responseMap = map[string]string{
	"ResponseOptions": ResponseOptions,
	"ResponseReqmod1": ResponseReqmod1,
	"ResponseReqmod2": ResponseReqmod2,
	"ResponseReqmod3": ResponseReqmod3,
	"ResponseRespmod": ResponseRespmodPart1 + ResponseRespmodPart2,
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	var data string
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}
		s := string(buf[:n])
		if strings.Contains(s, CRLF) { //read until ending marker
			data += strings.TrimRight(s, CRLF)
			break
		}
		data += s
	}
	if _, ok := responseMap[data]; ok {
		response := []byte(responseMap[data])
		quarter := int(len(response) / 4)
		conn.Write(response[:quarter])
		time.Sleep(10 * time.Millisecond)
		conn.Write(response[quarter:2*quarter])
		time.Sleep(10 * time.Millisecond)
		conn.Write(response[2*quarter:3*quarter])
		time.Sleep(10 * time.Millisecond)
		conn.Write(response[3*quarter:])
	}
	conn.Close()
}

func startServer(host string, port int, handler func(net.Conn)) net.Listener {
	addr := fmt.Sprintf(host+":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Server stopped")
				break
			}
			go handler(conn)
		}
	}()
	return ln
}

func getDriver(t *testing.T, host string, port int) *Driver {
	driver := &Driver{
		Host:          host,
		Port:          port,
		DialerTimeout: time.Second,
		ReadTimeout:   time.Second,
		WriteTimeout:  time.Second,
	}

	var err error
	for i := 1; i <= 5; i++ {
		err := driver.Connect()
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Errorf("Driver connection failed: %s", err.Error())
	}
	return driver
}

func checkResponse(t *testing.T, host string, port int, name string, valid bool) {
	driver := getDriver(t, host, port)
	err := driver.Send([]byte(name + CRLF))
	if err != nil {
		t.Errorf("Error sending data to server: %s", err.Error())
	}

	msg, err := driver.tcp.read()
	if !valid { // response was invalid
		if err == nil {
			t.Error("Parsed incorrect response without error!")
		}
		return
	}
	// response was valid
	if err != nil {
		t.Errorf("Received error for %s: %v", name, err)
	}

	assert.Equal(t, responseMap[name], msg)
	if err := driver.Close(); err != nil {
		t.Errorf("Driver connection close failed: %s", err.Error())
	}
}

// TestResponseParsing checks that all typical responses from ICAP server can be read by transport
// correctly without producing errors
func TestResponseParsing(t *testing.T) {
	ln := startServer("127.0.0.1", 13445, handleRequest)

	for name := range responseMap {
		checkResponse(t, "127.0.0.1", 13445, name, true)
	}
	checkResponse(t, "127.0.0.1", 13445, "NotValidResponse", false)
	ln.Close()
}

// handlerSectionTest returns ResponseRespmod divided into 2 parts by DoubleCRLF
// (DoubleCRLF without other conditions was mistakenly used as a mark of message ending)
func handlerSectionTest(conn net.Conn) {
	conn.Write([]byte(ResponseRespmodPart1))
	time.Sleep(10 * time.Millisecond)
	conn.Write([]byte(ResponseRespmodPart2))
}

// TestAllSectionsProcessed checks that transport:read() doesn't stop processing at DoubleCRLF
func TestAllSectionsProcessed(t *testing.T) {
	ln := startServer("127.0.0.1", 13446, handlerSectionTest)
	checkResponse(t, "127.0.0.1", 13446, "ResponseRespmod", true)
	ln.Close()
}

// Test findLastSectionStart
func TestFindLastSectionStart(t *testing.T) {
	assert.Equal(t, 0, findLastSectionStart([]byte(ResponseOptions)))
	assert.Equal(t, 231, findLastSectionStart([]byte(ResponseReqmod1)))
	assert.Equal(t, 244, findLastSectionStart([]byte(ResponseReqmod2)))
	assert.Equal(t, 213, findLastSectionStart([]byte(ResponseReqmod3)))
	assert.Equal(t, 222, findLastSectionStart([]byte(ResponseRespmodPart1)))
}
