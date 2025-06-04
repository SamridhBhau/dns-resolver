package message

import (
	"encoding/binary"
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

type ResourceRecord struct {
	Name string 
	Type uint16
	Class uint16
	TTL uint32
	RDLength uint16
	RData []byte
}

type Message struct {
	H Header
	Q Question
	ANS ResourceRecord
	AUTH ResourceRecord
	ADD ResourceRecord
}

func (q Question) EncodeName() ([]byte, error) {
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

func (h Header) Marshal() []byte {
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

func (q Question) Marshal() []byte {
	qName,_ := q.EncodeName()

	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, qName)
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.QType)
	byteArr = binary.BigEndian.AppendUint16(byteArr, q.QClass)
	return byteArr
}

func (m Message) Marshal() []byte{
	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, m.H.Marshal())
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, m.Q.Marshal())

	return byteArr
}
