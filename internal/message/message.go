package message

import (
	"encoding/binary"
	"fmt"
	"github.com/SamridhBhau/dnsResolver/internal/utils"
)

const UDPMAXSIZE = 512

type Header struct {
	ID      uint16
	QR      bool  // query(0) -> false, response(1) -> true
	OPCODE  uint8 // 0-15
	AA      bool
	TC      bool
	RD      bool
	RA      bool
	Z       uint8
	RCODE   uint8
	QDCOUNT uint16
	ANCOUNT uint16
	NSCOUNT uint16
	ARCOUNT uint16
}

type Question struct {
	QName  string
	QType  uint16
	QClass uint16
}

type ResourceRecord struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    string
}

type Message struct {
	H    Header
	Q    Question
	ANS  []ResourceRecord
	AUTH []ResourceRecord
	ADD  []ResourceRecord
}

func (h Header) Marshal() []byte {
	var byteArr []byte
	byteArr = binary.BigEndian.AppendUint16(byteArr, h.ID)

	// first 8 bits of flags
	var firstByte uint8 = 0

	// QR bit
	if h.QR == true {
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

func (q Question) Marshal() []byte {
	qName, _ := utils.EncodeName(q.QName)

	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, qName)
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.QType)
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.QClass)
	return byteArr
}

func (m Message) Marshal() []byte {
	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, m.H.Marshal())
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, m.Q.Marshal())

	return byteArr
}

func (h Header) Display() {
	fmt.Printf("ID: %d\n", h.ID)
	fmt.Printf("QR: %t\n", h.QR)
	fmt.Printf("OPCODE: %d\n", h.OPCODE)
	fmt.Printf("AA: %t\n", h.AA)
	fmt.Printf("TC: %t\n", h.TC)
	fmt.Printf("RD: %t\n", h.RD)
	fmt.Printf("RA: %t\n", h.RA)
	fmt.Printf("Z: %d\n", h.Z)
	fmt.Printf("RCODE: %d\n", h.RCODE)
	fmt.Printf("QDCOUNT: %d\n", h.QDCOUNT)
	fmt.Printf("ANCOUNT: %d\n", h.ANCOUNT)
	fmt.Printf("NSCOUNT: %d\n", h.NSCOUNT)
	fmt.Printf("ARCOUNT: %d\n", h.ARCOUNT)
}

func (q Question) Display() {
	fmt.Printf("QNAME: %s\n", q.QName)
	fmt.Printf("QTYPE: %d\n", q.QType)
	fmt.Printf("QCLASS: %d\n", q.QClass)
}

func (RR ResourceRecord) Display() {
	fmt.Printf("NAME: %s\n", RR.Name)
	fmt.Printf("TYPE: %d\n", RR.Type)
	fmt.Printf("CLASS: %d\n", RR.Class)
	fmt.Printf("TTL: %d\n", RR.TTL)
	fmt.Printf("RDLENGTH: %d\n", RR.RDLength)
	fmt.Printf("RDATA: %s\n", RR.RData)
}

func (m Message) Display() {
	m.H.Display()
	m.Q.Display()

	for _, ans := range m.ANS {
		ans.Display()
	}

	for _, auth := range m.AUTH {
		auth.Display()
	}

	for _, add := range m.ADD {
		add.Display()
	}
}
