package main

import (
	"fmt"
	"github.com/SamridhBhau/dnsResolver/internal/message"
	"github.com/SamridhBhau/dnsResolver/internal/parse"
	"net"
	"os"
)

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

	message := message.Message{
		H: header,
		Q: question,
	}

	msgBytes := message.Marshal()

	udpAddr, err := net.ResolveUDPAddr("udp", "8.8.8.8:53")
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

	recvBuf := make([]byte, 1024)
	_, err = conn.Read(recvBuf)

	if err != nil {
		fmt.Println("Read failed: ", err.Error())
		os.Exit(1)
	}

	resMsg := parse.ParseResponse(recvBuf)
	resMsg.Display()

}
