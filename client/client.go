package main

import (
	"encoding/binary"
	"errors"
	"flag"
	cfg "go-proxy/config"
	"go-proxy/tool"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	configPath       = flag.String("conf", "conf.yml", "configuration file path")
	conf             cfg.Config
	disabledRedirect = errors.New("disabled redirect.")
)

type TRPClient struct {
	svrAddr string
	maxConn int
	token   string
	sync.Mutex
}

func NewRPClient(svraddr string, maxConn int, token string) *TRPClient {
	c := new(TRPClient)
	c.svrAddr = svraddr
	c.maxConn = maxConn
	c.token = token
	return c
}

func (c *TRPClient) Start() error {
	for i := 0; i < c.maxConn; i++ {
		go c.newConn()
	}
	for {
		time.Sleep(5 * time.Second)
	}
}

func (c *TRPClient) newConn() error {
	c.Lock()
	conn, err := net.Dial("tcp", c.svrAddr)
	if err != nil {
		log.Println("Failed to connect to the server and will reconnect in five seconds.")
		time.Sleep(time.Second * 5)
		c.Unlock()
		_ = c.newConn()
		return err
	}
	c.Unlock()
	_ = conn.(*net.TCPConn).SetKeepAlive(true)
	//conn.(*net.TCPConn).SetKeepAlivePeriod(time.Duration(2 * time.Second))
	return c.process(conn)
}

func (c *TRPClient) werror(conn net.Conn) {
	_, _ = conn.Write([]byte(cfg.SERVER_ERROR))
}

func (c *TRPClient) process(conn net.Conn) error {
	if _, err := conn.Write(tool.CheckValue(c.token)); err != nil {
		return err
	}
	val := make([]byte, 4)
	for {
		_, err := conn.Read(val)
		if err != nil {
			log.Printf("The server is disconnected with error %s and will reconnect in five seconds.\n", err.Error())
			time.Sleep(5 * time.Second)
			go c.newConn()
			return err
		}
		flags := string(val)
		switch flags {
		case cfg.VERIFY_FAILED:
			log.Fatal("Token is incorrect.")
		case cfg.VERIFY_SUCCESSED:
			_ = c.deal(conn)
		case cfg.SERVER_ERROR:
			log.Println("The server returned an error.")
		default:
			log.Println("The error cannot be resolved.")
		}
	}
}

func (c *TRPClient) deal(conn net.Conn) error {
	val := make([]byte, 4)
	_, err := conn.Read(val)
	if err != nil {
		log.Println(err)
	}
	nlen := binary.LittleEndian.Uint32(val)
	log.Println("Received server data length: ", nlen)
	if nlen <= 0 {
		c.werror(conn)
		return errors.New("Data length error")
	}
	raw := make([]byte, nlen)
	n, err := conn.Read(raw)
	if err != nil {
		return err
	}
	if n != int(nlen) {
		log.Printf("The length of the data on the read server is wrong, it has been read %d byte, the total byte length is %d byte\n", n, nlen)
		c.werror(conn)
		return errors.New("Read server data length error.")
	}
	req, err := tool.DecodeRequest(raw, conf.Tunnel)
	if err != nil {
		log.Println("DecodeRequest error, ", err)
		c.werror(conn)
		return err
	}
	rawQuery := ""
	if req.URL.RawQuery != "" {
		rawQuery = "?" + req.URL.RawQuery
	}
	log.Println(req.URL.Path + rawQuery)
	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return disabledRedirect
	}
	resp, err := client.Do(req)
	disRedirect := err != nil && strings.Contains(err.Error(), disabledRedirect.Error())
	if err != nil && !disRedirect {
		log.Println("Request local client error, ", err)
		c.werror(conn)
		return err
	}
	if !disRedirect {
		defer resp.Body.Close()
	} else {
		resp.Body = nil
		resp.ContentLength = 0
	}
	respBytes, err := tool.EncodeResponse(resp, conf.DomainAsProto, conf.Tunnel)
	if err != nil {
		log.Printf("EncodeResponse error: %s\n", err.Error())
		c.werror(conn)
		return err
	}
	n, err = conn.Write(respBytes)
	if err != nil {
		log.Println("Sending data error, ", err)
		return err
	}
	if n != len(respBytes) {
		log.Printf("The length of the sent data is wrong and it has been sent % dbyte, the total byte length is %d byte\n", n, len(respBytes))
	} else {
		log.Printf("This request was successfully completed, a total of: %d byte\n", n)
	}
	return nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cfg.Init(*configPath, &conf)

	log.Printf("The client has started, ready to connect to %s\n", conf.ServerAddr)
	cli := NewRPClient(conf.ServerAddr, conf.MaxConn, conf.Token)
	_ = cli.Start()
}
