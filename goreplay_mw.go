package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/buger/goreplay/proto"
)

type old_to_new struct {
	old string
	new []string
}

// requestID -> originalToken
var sessionIDs map[string]old_to_new

// var cntr int = 0

func main() {
	sessionIDs = make(map[string]old_to_new)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		encoded := scanner.Bytes()
		buf := make([]byte, len(encoded)/2)
		hex.Decode(buf, encoded)

		process(buf)
	}
}

func get_session_id(ele []string) string {
	for _, v := range ele {
		if strings.Contains(v, "SESSION_ID") {
			clean_raw_v := strings.Split(v, ";")[0]
			return clean_raw_v
		}
	}
	return ""
}

func get_session_id_from_cookie(ele []string) string {
	for _, v := range ele {
		if strings.Contains(v, "SESSION_ID") {
			// Debug("::: SESSION_ID", v)
			for _, val := range strings.Split(v, ";") {
				if strings.Contains(val, "SESSION_ID") {
					return val
				}
			}

		}
	}
	return ""
}

func create_cookie_value_from_list(lst []string) string {
	raw := *new([]string)
	for _, v := range lst {
		// append(raw, strings.Split(v, ";")[0])
		raw = append(raw, strings.Split(v, ";")[0])
	}
	res := strings.Join(raw, "; ")
	return res
}

func process(buf []byte) {
	payloadType := buf[0]
	headerSize := bytes.IndexByte(buf, '\n') + 1
	header := buf[:headerSize-1]

	meta := bytes.Split(header, []byte(" "))
	reqID := string(meta[1])
	payload := buf[headerSize:]

	hs := proto.ParseHeaders(payload)

	// body := proto.Body(payload)

	switch payloadType {
	case '1':
		if _, ok := hs["Cookie"]; ok {
			if _, ok := sessionIDs[reqID]; !ok {
				sessionIDs[reqID] = *new(old_to_new)
			}

			ele := proto.Header(payload, []byte("Cookie"))
			resp := get_session_id_from_cookie([]string{string(ele)})

			Debug("ele: ", ele)
			Debug("resp: ", resp)

			// for _, val := range sessionIDs {
			// 	if strings.TrimSpace(val.old) == strings.TrimSpace(resp) {
			// 		new_cookie := create_cookie_value_from_list(val.new)

			// 		payload = proto.SetHeader(payload, []byte("Cookie"), []byte(new_cookie))
			// 		buf = append(buf[:headerSize], payload...)
			// 		// Debug("REQ", new_cookie, val.new)
			// 		os.Stdout.Write(encode(buf))
			// 		return
			// 	}
			// }
		}
		os.Stdout.Write(encode(buf))
		Debug("ORIG")
	case '2':
		if s_elem, ok := sessionIDs[reqID]; ok {
			if _, ok := hs["Set-Cookie"]; ok {
				for key, ele := range hs {
					if key == "Set-Cookie" {
						resp := get_session_id(ele)
						s_elem.old = resp
						sessionIDs[reqID] = s_elem
					}
				}
			}
		}
		// os.Stdout.Write(encode(buf))
		Debug("RESP")
	case '3':
		// status := string(proto.Status(payload))
		if s_elem, ok := sessionIDs[reqID]; ok {
			if _, ok := hs["Set-Cookie"]; ok {
				Debug("REP REQID", sessionIDs[reqID])
				for key, ele := range hs {
					if key == "Set-Cookie" {
						s_elem.new = []string(ele)
						// Debug("REP", sessionIDs[reqID])
						sessionIDs[reqID] = s_elem
					}
				}
			}
		}
		// Debug("::> REPLAY STATUS: ", status)
		Debug("REPLAY")
	}
}

// --------------------------------------------------------------------------
func Debug(args ...interface{}) {
	if os.Getenv("GOR_TEST") == "" {
		fmt.Fprint(os.Stderr, "[DEBUG] [sid-MOD] ")
		fmt.Fprintln(os.Stderr, args...)
	}
}

func encode(buf []byte) []byte {
	dst := make([]byte, len(buf)*2+1)
	hex.Encode(dst, buf)
	dst[len(dst)-1] = '\n'

	return dst
}
