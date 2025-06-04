package parse

import (
	"github.com/SamridhBhau/dnsResolver/internal/message"
	"encoding/binary"
)

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

func DecodeName(name []byte) string {
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

func ParseQuestion(msg []byte) message.Question{
	name := DecodeName(msg)

	i := len(name)+2
	qtype := binary.BigEndian.Uint16(msg[i:i+2])
	i += 2
	qclass := binary.BigEndian.Uint16(msg[i:i+2])

	return message.Question{
		QName : name,
		QType : qtype,
		QClass : qclass,
	}
}

func ParseResponse(msg []byte) message.Message{
	response := message.Message{}
	response.H = ParseHeader(msg)
	response.Q = ParseQuestion(msg[12:])
	//TODO: 
	/*
  response.ans = ParseAnswer(message)
	response.auth = ParseAuthorities(message)
	response.add = ParseAdditional(message)
  */

	return response
}
