package parse

import (
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/SamridhBhau/dnsResolver/internal/message"
	"github.com/SamridhBhau/dnsResolver/internal/utils"
)

func ParseHeader(msg []byte) message.Header {
	// ID
	id := binary.BigEndian.Uint16(msg[:2])

	// Flags
	var firstByte byte = msg[2]
	qr := utils.IsSet(firstByte, 0)
	// four bits
	var opcode uint8 = (firstByte >> 3) & 0x0f
	aa := utils.IsSet(firstByte, 5)
	tc := utils.IsSet(firstByte, 6)
	rd := utils.IsSet(firstByte, 7)

	var secondByte = msg[3]
	ra := utils.IsSet(secondByte, 0)
	// 3 bits
	z := (secondByte >> 4) & 0x07
	// 4 bits
	var rcode uint8 = secondByte & (0x0f)

	// Counts
	qdcount := binary.BigEndian.Uint16(msg[4:6])
	ancount := binary.BigEndian.Uint16(msg[6:8])
	nscount := binary.BigEndian.Uint16(msg[8:10])
	arcount := binary.BigEndian.Uint16(msg[10:12])

	return message.Header{
		ID:      id,
		QR:      qr,
		OPCODE:  opcode,
		AA:      aa,
		TC:      tc,
		RD:      rd,
		RA:      ra,
		Z:       z,
		RCODE:   rcode,
		QDCOUNT: qdcount,
		ANCOUNT: ancount,
		NSCOUNT: nscount,
		ARCOUNT: arcount,
	}
}

func ParseQuestion(msg []byte, start uint) (message.Question, uint) {
	name, bytesUsed := utils.DecodeName(msg, start)

	i := start + bytesUsed
	qtype := binary.BigEndian.Uint16(msg[i : i+2])
	i += 2
	qclass := binary.BigEndian.Uint16(msg[i : i+2])

	return message.Question{
		QName:  name,
		QType:  qtype,
		QClass: qclass,
	}, i + 2 - start
}

func ParseRdata(msg []byte, start uint, msgType uint16, rdlength uint16) string {
	switch msgType {
	case 1:
		var strs []string
		for i := uint(0); i < uint(rdlength); i++ {
			n := int(msg[start+i])
			strs = append(strs, strconv.Itoa(n))
		}
		return strings.Join(strs, ".")
	case 2:
		name, _ := utils.DecodeName(msg, start)
		return name
	default:
		return ""
	}
}

func ParseRR(msg []byte, start uint) (message.ResourceRecord, uint) {
	name, bytesUsed := utils.DecodeName(msg, start)
	i := start + bytesUsed

	AType := binary.BigEndian.Uint16(msg[i : i+2])
	i += 2
	class := binary.BigEndian.Uint16(msg[i : i+2])
	i += 2
	ttl := binary.BigEndian.Uint32(msg[i : i+4])
	i += 4
	rdlength := binary.BigEndian.Uint16(msg[i : i+2])
	i += 2

	rdata := ParseRdata(msg, i, AType, rdlength)
	i += uint(rdlength)

	return message.ResourceRecord{
		Name:     name,
		Type:     AType,
		Class:    class,
		TTL:      ttl,
		RDLength: rdlength,
		RData:    rdata,
	}, i - start
}

func ParseResponse(msg []byte) message.Message {
	header := ParseHeader(msg)
	var totalBytes uint = 12
	question, bytesUsed := ParseQuestion(msg, totalBytes)
	totalBytes += bytesUsed

	var ans []message.ResourceRecord
	for i := uint16(0); i < header.ANCOUNT; i++ {
		rr, bytesUsed := ParseRR(msg, totalBytes)
		totalBytes += bytesUsed

		ans = append(ans, rr)
	}

	var auth []message.ResourceRecord
	for i := uint16(0); i < header.NSCOUNT; i++ {
		rr, bytesUsed := ParseRR(msg, totalBytes)
		totalBytes += bytesUsed

		auth = append(auth, rr)
	}

	var add []message.ResourceRecord
	for i := uint16(0); i < header.ARCOUNT; i++ {
		rr, bytesUsed := ParseRR(msg, totalBytes)
		totalBytes += bytesUsed

		add = append(add, rr)
	}

	return message.Message{
		H:    header,
		Q:    question,
		ANS:  ans,
		AUTH: auth,
		ADD:  add,
	}
}
