package main

import (
	"fmt"
	"os"

	"github.com/SamridhBhau/dnsResolver/internal/message"
)

const RootNameServer = "192.33.4.12:53"

func main() {
	header := message.Header{
		ID:      22,
		QDCOUNT: 1,
	}

	question := message.Question{
		QName:  "news.ycombinator.com",
		QType:  1,
		QClass: 1,
	}

	msg := message.Message{
		H: header,
		Q: question,
	}

	ip, err := msg.Resolve(RootNameServer)
	if err != nil {
		os.Exit(1)
	}
	fmt.Println("IP address is: " + ip)
}
