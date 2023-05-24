package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	cfg "go-proxy/config"
	"go-proxy/tool"
	"io"

	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type TRPServer struct {
	tcpPort  int
	httpPort int
	token    string
	listener *net.TCPListener
	connList chan net.Conn
	sync.RWMutex
}

func NewRPServer(tcpPort, httpPort int, token string) *TRPServer {
	s := new(TRPServer)
	s.tcpPort = tcpPort
	s.httpPort = httpPort
	s.token = token
	s.connList = make(chan net.Conn, 1000)
	return s
}

func (s *TRPServer) Start() error {
	var err error
	s.listener, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: s.tcpPort, Zone: ""})
	if err != nil {
		return err
	}
	go s.httpserver()
	return s.tcpserver()
}

func (s *TRPServer) Close() error {
	if s.listener != nil {
		err := s.listener.Close()
		s.listener = nil
		return err
	}
	return errors.New("TCP instance is not created!")
}

func (s *TRPServer) tcpserver() error {
	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.cliProcess(conn)
	}
}

func badRequest(w http.ResponseWriter) {
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func (s *TRPServer) httpserver() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	retry:
		if len(s.connList) == 0 {
			badRequest(w)
			return
		}
		conn := <-s.connList
		log.Println(r.RequestURI)
		err := s.write(r, conn)
		if err != nil {
			log.Println(err)
			conn.Close()
			goto retry
		}
		err = s.read(w, conn)
		if err != nil {
			log.Println(err)
			conn.Close()
			goto retry
		}
		s.connList <- conn
		conn = nil
	})
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", s.httpPort), nil))
}

func (s *TRPServer) cliProcess(conn *net.TCPConn) error {
	_ = conn.SetReadDeadline(time.Now().Add(time.Duration(5) * time.Second))
	vval := make([]byte, 20)
	_, err := conn.Read(vval)
	if err != nil {
		log.Printf("Client %s read timeout.\n", conn.RemoteAddr())
		conn.Close()
		return err
	}
	if !bytes.Equal(vval, tool.CheckValue(s.token)[:]) {
		log.Printf("The connection of client %s verified error, will to be closed.\n", conn.RemoteAddr())
		_, _ = conn.Write([]byte(cfg.VERIFY_FAILED))
		conn.Close()
		return err
	}
	_ = conn.SetReadDeadline(time.Time{})
	log.Printf("Connect the new client: %s\n", conn.RemoteAddr())
	_ = conn.SetKeepAlive(true)
	_ = conn.SetKeepAlivePeriod(time.Duration(2 * time.Second))
	s.connList <- conn
	return nil
}

func (s *TRPServer) write(r *http.Request, conn net.Conn) error {
	raw, err := tool.EncodeRequest(r)
	if err != nil {
		return err
	}
	c, err := conn.Write(raw)
	if err != nil {
		return err
	}
	if c != len(raw) {
		return errors.New("The written length is inconsistent with the byte length.")
	}
	return nil
}

func (s *TRPServer) read(w http.ResponseWriter, conn net.Conn) error {
	val := make([]byte, 4)
	_, err := conn.Read(val)
	if err != nil {
		return err
	}
	flags := string(val)
	switch flags {
	case cfg.VERIFY_SUCCESSED:
		_, err = conn.Read(val)
		if err != nil {
			return err
		}
		nlen := int(binary.LittleEndian.Uint32(val))
		if nlen == 0 {
			return errors.New("Reading client length error.")
		}
		log.Printf("Receive client data, need to read length: %d\n", nlen)
		raw := make([]byte, 0)
		buff := make([]byte, 1024)
		c := 0
		for {
			clen, err := conn.Read(buff)
			if err != nil && err != io.EOF {
				return err
			}
			raw = append(raw, buff[:clen]...)
			c += clen
			if c >= nlen {
				break
			}
		}
		log.Printf("Reading completed, length: %d, actual raw length: %d\n", c, len(raw))
		if c != nlen {
			return fmt.Errorf("Read error, has been read %d byte, need to read %d byte.", c, nlen)
		}
		resp, err := tool.DecodeResponse(raw)
		if err != nil {
			return err
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		for k, v := range resp.Header {
			for _, v2 := range v {
				w.Header().Set(k, v2)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(bodyBytes)
	case cfg.SERVER_ERROR:
		return nil
	default:
		log.Printf("Unable to resolve the error: %s\n", val)
	}
	return nil
}

var (
	token    = flag.String("token", "88888888", "Verification key")
	logPath  = flag.String("log", "proxy.log", "Log file path")
	tcpPort  = flag.Int("tcpport", 8284, "Socket tcp connection port")
	httpPort = flag.Int("httpport", 8024, "Http connection port")
)

func main() {
	flag.Parse()
	f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(f)

	if *tcpPort <= 0 || *tcpPort >= 65536 {
		log.Fatalln("Incorrect tcp port.")
	}
	if *httpPort <= 0 || *httpPort >= 65536 {
		log.Fatalln("Incorrect http port.")
	}
	log.Printf("The server starts and listens to the tcp port: %d, http port: %d\n", *tcpPort, *httpPort)
	svr := NewRPServer(*tcpPort, *httpPort, *token)
	if err := svr.Start(); err != nil {
		log.Fatalln(err)
	}
	defer svr.Close()
}
