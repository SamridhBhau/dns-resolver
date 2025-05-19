package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

type Header struct {
	ID uint16 
	QR bool 
	OPCODE uint8 // 0-15
	AA bool
	TC bool
	RD bool
	RA bool
	Z uint8
	RCOUNT uint8
	QDCOUNT uint16
	ANCOUNT uint16
	NSCOUNT uint16
	ARCOUNT uint16
}

type Question struct {
	qName string
	qType uint16
	qClass uint16
}

func (q Question) NameToBytes() ([]byte, error) {
	var byteArr []byte
	substrs := strings.Split(q.qName, ".");

	for _, str := range substrs {
		l := uint64(len(str))

		byteArr = binary.AppendUvarint(byteArr, l)
		byteArr, _ = binary.Append(byteArr, binary.BigEndian, []byte(str))
	}
	byteArr = binary.AppendUvarint(byteArr, 0)
	return byteArr, nil
}

func (h Header) ConvertToBytes() []byte {
	var byteArr []byte
	byteArr = binary.BigEndian.AppendUint16(byteArr, h.ID)

	// first 8 bits of flags
	var firstByte uint8 = 0

	// QR bit
	if h.QR == true{
		firstByte |= (1 << 7)
	}

	// OPCODE - 4 bits
	firstByte |= (h.OPCODE << 3)

	// AA bit
	if h.AA == true {
		firstByte |= (1 << 2)
	}

	// TC bit
	if h.TC == true {
		firstByte |= (1 << 1)
	}

	// RD bit
	if h.RD == true {
		firstByte |= 1 
	}

	var secondByte uint8 = 0

	// RA bit
	if h.RA == true {
		secondByte |= (1 << 7)
	}

	// Z field - 3 bits - reserved

	// RCODE - 4 bits
	secondByte |= h.RCOUNT

	var flags uint16 = 0
	flags |= (uint16(firstByte) << 8)
	flags |= uint16(secondByte)

	// Append flags
	byteArr = binary.BigEndian.AppendUint16(byteArr, flags)

	// Append count of question, answers, authority and additional
	byteArr = binary.BigEndian.AppendUint16(byteArr, h.QDCOUNT)
	byteArr = binary.BigEndian.AppendUint16(byteArr, h.ANCOUNT)
	byteArr = binary.BigEndian.AppendUint16(byteArr, h.NSCOUNT)
	byteArr = binary.BigEndian.AppendUint16(byteArr, h.ARCOUNT)

	return byteArr
}

func (q Question) ConvertToBytes() []byte {
	qName,_ := q.NameToBytes()

	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, qName)
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.qType)
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.qClass)
	return byteArr
}

func (m Message) ConvertToBytes() []byte{
	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, m.h.ConvertToBytes())
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, m.q.ConvertToBytes())

	return byteArr
}

type Message struct {
	h Header
	q Question
}


func main() {
	header := Header{
		ID : 22,
		RD : true,
		QDCOUNT: 1,
	}

	question := Question {
		qName : "dns.google.com",
		qType : 1,
		qClass : 1,
	}

	message := Message {
		h : header,
		q : question,
	}

	msgBytes := message.ConvertToBytes()

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

	fmt.Println(recvBuf)
}
