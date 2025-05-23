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
	QR bool  // query(0) -> false, response(1) -> true
	OPCODE uint8 // 0-15
	AA bool
	TC bool
	RD bool
	RA bool
	Z uint8
	RCODE uint8
	QDCOUNT uint16
	ANCOUNT uint16
	NSCOUNT uint16
	ARCOUNT uint16
}

type Question struct {
	QName string
	QType uint16
	QClass uint16
}
//TODO: Resource Record format

func (q Question) NameToBytes() ([]byte, error) {
	var byteArr []byte
	substrs := strings.Split(q.QName, ".");

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
	secondByte |= h.RCODE

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
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.QType)
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.QClass)
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
	/*
	ans ResourceRecord
	auth ResourceRecord
	add ResourceRecord
*/
}


func main() {
	header := Header{
		ID : 22,
		RD : true,
		QDCOUNT: 1,
	}

	question := Question {
		QName : "dns.google.com",
		QType : 1,
		QClass : 1,
	}


	message := Message {
		h : header,
		q : question,
	}


	msgBytes := message.ConvertToBytes()
	fmt.Println(msgBytes)

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

// isSet checks if bit at pos is set or not. pos start from 0
func isSet(b byte, pos uint) bool{
	var mask byte = (1 << (7 - pos))

	if (b & mask) != 0{
		return true
	} 
	return false
	
}

func ParseHeader(message []byte) Header {
	header := Header{
	}

	// ID
	header.ID = binary.BigEndian.Uint16(message[:2])

	// Flags
	var firstByte byte = message[2]
	header.QR = isSet(firstByte, 0)
	// four bits
	header.OPCODE = firstByte & (0x0f << 3)
	header.AA = isSet(firstByte, 5)
	header.TC = isSet(firstByte, 6)
	header.RD = isSet(firstByte, 7)

	var secondByte = message[3]
	header.RA = isSet(secondByte, 0)
	// 3 bits
	header.Z = secondByte & (0x07 << 4)
	// 4 bits
	header.RCODE = secondByte & (0x0f)

	// Counts
	header.QDCOUNT = binary.BigEndian.Uint16(message[4:6])
	header.ANCOUNT = binary.BigEndian.Uint16(message[6:8])
	header.NSCOUNT = binary.BigEndian.Uint16(message[8:10])
	header.ANCOUNT = binary.BigEndian.Uint16(message[10:12])

	return header
}

func ByteToName(name []byte) string {
	res := make([]byte, len(name))
	copy(res, name)

	i := 0
	for i < len(res){
		n := int(res[i])
		if n == 0 {
			break
		}

		if i != 0 {
			res[i] = byte('.')
		}
		
		i += (n + 1)
	}

	str := string(res[1:i])

	return str 
}

func ParseQuestion(message []byte) Question{
	name := ByteToName(message)

	i := len(name)+2
	qtype := binary.BigEndian.Uint16(message[i:i+2])
	i += 2
	qclass := binary.BigEndian.Uint16(message[i:i+2])

	return Question{
		QName : name,
		QType : qtype,
		QClass : qclass,
	}
}

func ParseResponse(message []byte) Message{
	response := Message{}
	response.h = ParseHeader(message)
	response.q = ParseQuestion(message[12:])
	//TODO: 
	/*
  response.ans = ParseAnswer(message)
	response.auth = ParseAuthorities(message)
	response.add = ParseAdditional(message)
  */

	return response
}
