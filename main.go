package main

import (
	"fmt"
	"github.com/SamridhBhau/dnsResolver/internal/message"
	"github.com/SamridhBhau/dnsResolver/internal/parse"
	"net"
	"os"
)

const RootNameServer = "198.41.0.4:53"

func main() {
	header := message.Header{
		ID:      22,
		RD:      true,
		QDCOUNT: 1,
	}

	question := message.Question{
		QName:  "example.com",
		QType:  1,
		QClass: 1,
	}

	msg := message.Message{
		H: header,
		Q: question,
	}

	msgBytes := msg.Marshal()

	udpAddr, err := net.ResolveUDPAddr("udp", RootNameServer)
	if err != nil {
		fmt.Println("ResolveUDPAddr error", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("Listen failed: ", err.Error())
		os.Exit(1)
	}

	defer conn.Close()

	_, err = conn.Write(msgBytes)

	if err != nil {
		fmt.Println("Write failed: ", err.Error())
		os.Exit(1)
	}

	recvBuf := make([]byte, message.UDPMAXSIZE)
	_, err = conn.Read(recvBuf)

	if err != nil {
		fmt.Println("Read failed: ", err.Error())
		os.Exit(1)
	}

	resMsg := parse.ParseResponse(recvBuf)
	resMsg.Display()
}
