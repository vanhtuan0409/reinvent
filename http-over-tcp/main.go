package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
)

type myReponseWriter struct {
	conn   net.Conn
	body   *bytes.Buffer
	status int
	header http.Header
}

func (rw *myReponseWriter) Header() http.Header {
	return rw.header
}

func (rw *myReponseWriter) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *myReponseWriter) WriteHeader(status int) {
	rw.status = status
}

func (rw *myReponseWriter) flush() error {
	status := fmt.Sprintf("HTTP/1.1 %d\n", rw.status)
	if _, err := rw.conn.Write([]byte(status)); err != nil {
		return err
	}

	body := rw.body.Bytes()
	rw.Header().Set("Content-Length", strconv.Itoa(len(body)))
	if err := rw.Header().Write(rw.conn); err != nil {
		return err
	}
	rw.conn.Write([]byte("\n"))
	if _, err := rw.conn.Write(rw.body.Bytes()); err != nil {
		return err
	}
	return nil
}

func newResponseWriter(conn net.Conn) *myReponseWriter {
	rw := new(myReponseWriter)
	rw.conn = conn
	buf := []byte{}
	rw.body = bytes.NewBuffer(buf)
	rw.header = http.Header{}
	return rw
}

func handleFunc(r *http.Request, w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<html><body><h1>Hello</h1></body></html>"))
}

func handleConnection(conn net.Conn) error {
	r := bufio.NewReader(conn)
	req, err := http.ReadRequest(r)
	if err != nil {
		log.Printf("Cannot parse request. ERR: %v\n", err)
		return err
	}
	message, err := httputil.DumpRequest(req, false)
	if err != nil {
		log.Printf("Cannot dump request. ERR: %v\n", err)
		return err
	}
	fmt.Println(string(message))

	writer := newResponseWriter(conn)
	handleFunc(req, writer)
	return writer.flush()
}

func main() {
	fmt.Println("Starting server")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Cannot accept connection. ERR: %v\n", err)
			continue
		}

		handleConnection(conn)
		conn.Close()
	}
}
