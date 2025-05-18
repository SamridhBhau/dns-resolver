package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

const ID uint16 = 22
type Header struct {
	id 		uint16 
	flags uint16
	qdCount uint16
	anCount uint16
	nsCount uint16
	arCount uint16
}

type Question struct {
	qName []byte
	qType uint16
	qClass uint16
}

func processString(s string) ([]byte, error) {
	var byteArr []byte
	substrs := strings.Split(s, ".");

	for _, str := range substrs {
		l := len(str)

		byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(l))

		byteArr, _ = binary.Append(byteArr, binary.BigEndian, []byte(str))
	}
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(0))
	return byteArr, nil
}

func createQuestion(s string) (Question, error) {
	byteArr, err := processString(s);
	if (err != nil) {
		return Question{}, errors.New("process string error")
	}

	return Question{
		qName: byteArr, 
		qType: 1,
		qClass : 1,
	}, nil
}

func createHeader() Header {
	return Header{
		id : ID,
		flags : 256,
		qdCount: 1,
	}
}

func (h Header) convertBytes() []byte {
	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(h.id))
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(h.flags))
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(h.qdCount))
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(h.anCount))
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(h.nsCount))
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(h.arCount))

	return byteArr
}

func (q Question) convertBytes() []byte {
	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, q.qName)
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(q.qType))
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, byte(q.qClass))

	return byteArr
}

type message struct {
	h Header
	q Question
}

func main() {
	question, _ := createQuestion("dns.google.com")
	header := createHeader();

	msg := message{
		header,
		question,
	}

	var byteArr []byte
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, msg.h.convertBytes())
	byteArr, _ = binary.Append(byteArr, binary.BigEndian, msg.q.convertBytes())

	for _, v := range byteArr {
		fmt.Printf("%x ", v)
	}
}
