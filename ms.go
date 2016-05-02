package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
)

var (
	host = flag.String("host", "0.0.0.0", "host to listen on")
	port = flag.Int("port", 27010, "port to listen on")

	blockSize = flag.Int("size", 1024, "block size to read packets on")

	file = flag.String("file", "servers", "file with server list")
)

func main() {
	flag.Parse()
	ip := net.ParseIP(*host)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: *port})
	if err != nil {
		fmt.Println(err)
		return
	}

	serverlist, err := os.Open(*file)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer serverlist.Close()

	reader := bufio.NewReader(serverlist)
	scanner := bufio.NewScanner(reader)

	data := make([]byte, *blockSize)

	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil {
			fmt.Printf("error during read: %s", err)
			return
		}

		buf := new(bytes.Buffer)

		binary.Write(buf, binary.LittleEndian, []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x66, 0x0A})

		for scanner.Scan() {
			host, port, err := net.SplitHostPort(scanner.Text())
			if err != nil {
				fmt.Println(err)
				return
			}
			ip = net.ParseIP(host).To4()
			if ip == nil {
				fmt.Printf("%v is not an IP address\n", ip)
				return
			}
			port_i, _ := strconv.Atoi(port)
			port_i16 := int16(port_i)
			port_o := port_i16<<8 | port_i16>>8

			binary.Write(buf, binary.LittleEndian, ip)
			binary.Write(buf, binary.LittleEndian, port_o)
		}

		binary.Write(buf, binary.LittleEndian, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

		fmt.Printf("ip: % x\n", buf.Bytes())

		_, err = listener.WriteToUDP(buf.Bytes(), remoteAddr)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("<%s> %s\n", remoteAddr, data[:n])
	}
}