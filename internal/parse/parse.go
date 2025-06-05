package parse

import (
	"encoding/binary"
	"github.com/SamridhBhau/dnsResolver/internal/message"
	"strings"
)

const MaxDomainLength = 255

// isSet checks if bit at pos is set or not. pos start from 0
func isSet(b byte, pos uint) bool {
	var mask byte = (1 << (7 - pos))

	if (b & mask) != 0 {
		return true
	}
	return false

}

func ParseHeader(msg []byte) message.Header {
	// ID
	id := binary.BigEndian.Uint16(msg[:2])

	// Flags
	var firstByte byte = msg[2]
	qr := isSet(firstByte, 0)
	// four bits
	var opcode uint8 = (firstByte >> 3) & 0x0f
	aa := isSet(firstByte, 5)
	tc := isSet(firstByte, 6)
	rd := isSet(firstByte, 7)

	var secondByte = msg[3]
	ra := isSet(secondByte, 0)
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

func DecodeName(msg []byte, start uint) (string, uint) {
	i := start
	var labels []string
	var jump bool
	var bytesUsed uint

	for i-start <= MaxDomainLength {
		// Check bits for pointer form
		if isSet(msg[i], 0) && isSet(msg[i], 1) {
			var offset uint16 = ((uint16(msg[i]) & 0x3F) << 8) | uint16(msg[i+1])
			start = uint(offset)
			i = start
			jump = true
			bytesUsed += 2
		}

		n := uint(msg[i])
		if n == 0 {
			break
		}

		labelStart := i + 1
		labelEnd := i + n
		label := string(msg[labelStart : labelEnd+1])
		if jump == false {
			bytesUsed += uint(len(label)) + 1
		}
		labels = append(labels, label)

		i += (n + 1)
	}

	// Add last length octet if not jumped
	if jump == false {
		bytesUsed++
	}

	return strings.Join(labels, "."), bytesUsed
}

func ParseQuestion(msg []byte, start uint) (message.Question, uint) {
	name, bytesUsed := DecodeName(msg, start)

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

func ParseRR(msg []byte, start uint) (message.ResourceRecord, uint) {
	name, bytesUsed := DecodeName(msg, start)
	i := start + bytesUsed

	AType := binary.BigEndian.Uint16(msg[i : i+2])
	i += 2
	class := binary.BigEndian.Uint16(msg[i : i+2])
	i += 2
	ttl := binary.BigEndian.Uint32(msg[i : i+4])
	i += 4
	rdlength := binary.BigEndian.Uint16(msg[i : i+2])
	i += 2

	rdata := make([]byte, rdlength)
	copy(rdata, msg[i:i+uint(rdlength)+1])
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
