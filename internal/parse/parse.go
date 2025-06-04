package parse

import (
	"encoding/binary"
	"strings"

	"github.com/SamridhBhau/dnsResolver/internal/message"
)

const MaxDomainLength = 255

// isSet checks if bit at pos is set or not. pos start from 0
func isSet(b byte, pos uint) bool{
	var mask byte = (1 << (7 - pos))

	if (b & mask) != 0{
		return true
	} 
	return false
	
}

func ParseHeader(msg []byte) message.Header {
	header := message.Header{
	}

	// ID
	header.ID = binary.BigEndian.Uint16(msg[:2])

	// Flags
	var firstByte byte = msg[2]
	header.QR = isSet(firstByte, 0)
	// four bits
	header.OPCODE = firstByte & (0x0f << 3)
	header.AA = isSet(firstByte, 5)
	header.TC = isSet(firstByte, 6)
	header.RD = isSet(firstByte, 7)

	var secondByte = msg[3]
	header.RA = isSet(secondByte, 0)
	// 3 bits
	header.Z = secondByte & (0x07 << 4)
	// 4 bits
	header.RCODE = secondByte & (0x0f)

	// Counts
	header.QDCOUNT = binary.BigEndian.Uint16(msg[4:6])
	header.ANCOUNT = binary.BigEndian.Uint16(msg[6:8])
	header.NSCOUNT = binary.BigEndian.Uint16(msg[8:10])
	header.ANCOUNT = binary.BigEndian.Uint16(msg[10:12])

	return header
}

func DecodeName(msg []byte, start uint) string {
	i := start
	var labels []string

	for i - start <= MaxDomainLength{
		// Check bits for pointer form
		if isSet(msg[i], 0) && isSet(msg[i], 1){
			var offset uint16 = ((uint16(msg[i]) & 0x3F) << 8) | uint16(msg[i+1])
			start = uint(offset)
			i = start
		}

		n := uint(msg[i])

		if n == 0{
			break
		}

		labelStart := i + 1
		labelEnd := i + n
		label := string(msg[labelStart: labelEnd+1])

		labels = append(labels, label)

		i += (n + 1)
	} 

	return strings.Join(labels, ".")
}

func ParseQuestion(msg []byte) (message.Question, uint){
	name := DecodeName(msg, 0)

	i := uint(len(name)+2)
	qtype := binary.BigEndian.Uint16(msg[i:i+2])
	i += 2
	qclass := binary.BigEndian.Uint16(msg[i:i+2])

	return message.Question{
		QName : name,
		QType : qtype,
		QClass : qclass,
	}, i + 1
}

func ParseAnswer(msg []byte, start uint) (message.ResourceRecord, uint){
	name := DecodeName(msg, 0)

	i := uint(len(name)+2)
	AType := binary.BigEndian.Uint16(msg[i:i+2])
	i += 2
	class := binary.BigEndian.Uint16(msg[i:i+2])
	i += 2
	ttl := binary.BigEndian.Uint32(msg[i:i+4])
	i += 4
	rdlength := binary.BigEndian.Uint16(msg[i:i+2])
	i += 2

	rdata := make([]byte, rdlength)
	copy(rdata, msg[i : i+uint(rdlength)+1])

	return message.ResourceRecord{
		Name : name,
		Type : AType, 
		Class : class,
		TTL : ttl,
		RDLength: rdlength,
		RData: rdata,
	}, i + 1
}

func ParseResponse(msg []byte) message.Message{
	header := ParseHeader(msg)
	question, bytesUsed := ParseQuestion(msg[12:])
	ans, bytesUsed := ParseAnswer(msg, 12+bytesUsed)

	//TODO: 
	/*
	response.auth = ParseAuthorities(message)
	response.add = ParseAdditional(message)
  */

	return message.Message{
		H : header,
		Q : question,
		ANS: ans,
	}
}
