package icapclient

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/egirna/icap"
)

var (
	port = 1344
)

const (
	previewBytes      = 24
	goodFileDetectStr = "GOOD FILE"
	badFileDetectStr  = "BAD FILE"
	badHost           = "badfile.com"
)

func startTestServer() {
	icap.HandleFunc("/respmod", respmodHandler)
	icap.HandleFunc("/reqmod", reqmodHandler)

	log.Println("Starting ICAP test server...")

	go func() {
		if err := icap.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			log.Println("Failed to start ICAP test server: ", err.Error())
			return
		}
	}()

	time.Sleep(5 * time.Millisecond)

	log.Printf("ICAP test server is running on localhost:%d\n...\n", port)
}

func respmodHandler(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", "ICAP-TEST")
	h.Set("Service", "ICAP-TEST-SERVICE")

	switch req.Method {
	case "OPTIONS":
		h.Set("Methods", "RESPMOD")
		h.Set("Allow", "204")
		if previewBytes > 0 {
			h.Set("Preview", strconv.Itoa(previewBytes))
		}
		h.Set("Transfer-Preview", "*")
		w.WriteHeader(http.StatusOK, nil, false)
	case "RESPMOD":
		defer req.Response.Body.Close()

		if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != "204") {
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// log.Println("The preview data: ", string(req.Preview))

		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, req.Response.Body); err != nil {
			log.Println("Failed to copy the response body to buffer: ", err.Error())
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		status := http.StatusNoContent

		if strings.Contains(buf.String(), badFileDetectStr) {
			status = http.StatusOK
		}

		w.WriteHeader(status, nil, false)

	}
}

func reqmodHandler(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", "ICAP-TEST")
	h.Set("Service", "ICAP-TEST-SERVICE")

	switch req.Method {
	case "OPTIONS":
		h.Set("Methods", "REQMOD")
		h.Set("Allow", "204")
		if previewBytes > 0 {
			h.Set("Preview", strconv.Itoa(previewBytes))
		}
		h.Set("Transfer-Preview", "*")
		w.WriteHeader(http.StatusOK, nil, false)
	case "REQMOD":

		if val, exist := req.Header["Allow"]; !exist || (len(val) > 0 && val[0] != "204") {
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// log.Println("The preview data: ", string(req.Preview))

		status := http.StatusNoContent

		if req.Request.Host == badHost {
			status = http.StatusOK
		}

		w.WriteHeader(status, nil, false)

	}
}

func testServerRunning() bool {
	lstnr, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return true
	}
	lstnr.Close()
	return false
}
