package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	socks5Ver = 0x05
	cmdBind   = 0x01
	atypeIPV4 = 0x01
	atypeHOST = 0x03
	atypeIPV6 = 0x04
)

func main() {
	server, err := net.Listen("tcp", "127.0.0.1:1080")
	if err != nil {
		panic(err)
	}
	for {
		client, err := server.Accept()
		if err != nil {
			log.Printf("Accept failed %v", err)
			continue
		}
		go process(client)
	}
}
func process(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	err := auth(reader, conn)
	if err != nil {
		log.Printf("client %v auth failed: %#v", conn.RemoteAddr(), err)
		return
	}
	err = connet(reader, conn)
	if err != nil {
		log.Printf("client %v connet failed: %#v", conn.RemoteAddr(), err)
		return
	}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		_, err = conn.Write([]byte{b})
		if err != nil {
			break
		}
	}
}
func auth(reader *bufio.Reader, conn net.Conn) (err error) {
	ver, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("read var failed: %v", err)
	}
	if ver != socks5Ver {
		return fmt.Errorf("not supported ver: %v", ver)
	}
	methodSize, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("read methodSize failed: %v", err)
	}
	method := make([]byte, methodSize)
	_, err = io.ReadFull(reader, method)
	if err != nil {
		return fmt.Errorf("read method failed: %v", err)
	}
	log.Println("ver", ver, "method", method)
	_, err = conn.Write([]byte{socks5Ver, 0x00})
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}
	return nil
}
func connet(reader *bufio.Reader, conn net.Conn) (err error) {
	buf := make([]byte, 4)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return fmt.Errorf("read header failed:%v", err)
	}
	ver, cmd, atype := buf[0], buf[1], buf[3]
	if ver != socks5Ver {
		return fmt.Errorf("not supported ver: %v", ver)
	}
	if cmd != cmdBind {
		return fmt.Errorf("not supported cmd: %v", cmd)
	}
	addr := ""
	switch atype {
	case atypeIPV4:
		_, err = io.ReadFull(reader, buf)
		if err != nil {
			return fmt.Errorf("read atype falied: %v", err)
		}
		addr = fmt.Sprintf("%d.%d.%d.%d", buf[0], buf[1], buf[2], buf[3])
	case atypeHOST:
		hostSize, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("read hostSize falied: %v", err)
		}
		host := make([]byte, hostSize)
		_, err = io.ReadFull(reader, host)
		if err != nil {
			return fmt.Errorf("read host falied: %v", err)
		}
		addr = string(host)
	case atypeIPV6:
		return errors.New("IPV6: not supported yet")
	default:
		return errors.New("invaild atype")
	}

	if ver != socks5Ver {
		return fmt.Errorf("not supported ver: %v", ver)
	}
	_, err = io.ReadFull(reader, buf[:2])
	if err != nil {
		return fmt.Errorf("read port failed: %v", err)
	}
	port := binary.BigEndian.Uint16(buf[:2])
	dest, err := net.Dial("tcp", fmt.Sprintf("%v:%v", addr, port))
	if err != nil {
		return fmt.Errorf("dial dest failed: %w", err)
	}
	defer dest.Close()
	log.Println("dial", addr, port)
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_, _ = io.Copy(dest, reader)
		cancel()
	}()
	go func() {
		_, _ = io.Copy(conn, dest)
		cancel()
	}()
	<-ctx.Done()
	return nil
}
