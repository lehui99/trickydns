package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	config := struct {
		BindAddr  string
		Servers   []string
		Timeout   time.Duration
		ExtraPtrs int
		BufSize   int
	}{}
	func() {
		buf, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(buf, &config); err != nil {
			panic(err)
		}
		config.Timeout *= time.Second
	}()
	udp := func() *net.UDPConn {
		bindAddr, err := net.ResolveUDPAddr("udp", config.BindAddr)
		if err != nil {
			panic(err)
		}
		udp, err := net.ListenUDP("udp", bindAddr)
		if err != nil {
			panic(err)
		}
		return udp
	}()
	bufPool := sync.Pool{New: func() interface{} {
		return make([]byte, config.BufSize)
	}}
	buf := bufPool.Get().([]byte)
	domain := bufPool.Get().([]byte)
	for {
		buf = buf[:cap(buf)]
		size, addr, err := udp.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		err = func() (err error) {
			defer func() {
				if recover() != nil {
					err = io.EOF
				}
			}()
			buf = buf[:size]
			for idx := 12; ; {
				if buf[idx] == 0 {
					domain = domain[:idx-11]
					break
				}
				if buf[idx] > 0x3f { // already has pointer or malformed
					return io.EOF
				}
				idx += int(buf[idx]) + 1
			}
			copy(domain, buf[12:])
			copy(buf[14:], buf[12+len(domain):])
			buf = buf[:len(buf)-len(domain)+2]
			binary.BigEndian.PutUint16(buf[12:], uint16(len(buf)))
			buf[12] |= 0xc0
			for i := 0; i < config.ExtraPtrs; i++ {
				buf = buf[:len(buf)+2]
				binary.BigEndian.PutUint16(buf[len(buf)-2:], uint16(len(buf)))
				buf[len(buf)-2] |= 0xc0
			}
			copy(buf[len(buf):cap(buf)], domain)
			buf = buf[:len(buf)+len(domain)]
			return nil
		}()
		if err != nil {
			continue
		}
		cli := func() *net.UDPConn {
			cliAddr, err := net.ResolveUDPAddr("udp", "")
			if err != nil {
				return nil
			}
			cli, err := net.ListenUDP("udp", cliAddr)
			if err != nil {
				return nil
			}
			for _, server := range config.Servers {
				serverAddr, err := net.ResolveUDPAddr("udp", server)
				if err != nil {
					continue
				}
				cli.WriteToUDP(buf, serverAddr)
			}
			return cli
		}()
		if cli == nil {
			continue
		}
		go func() {
			defer cli.Close()
			buf := bufPool.Get().([]byte)
			defer bufPool.Put(buf)
			cli.SetReadDeadline(time.Now().Add(config.Timeout))
			buf = buf[:cap(buf)]
			size, _, err := cli.ReadFromUDP(buf)
			if err != nil {
				return
			}
			buf = buf[:size]
			udp.WriteToUDP(buf, addr)
		}()
	}
}
