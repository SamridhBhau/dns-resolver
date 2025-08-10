package utils

import (
	"encoding/binary"
	"strings"
)

const MaxDomainNameLength = 255 * 2

func EncodeName(QName string) ([]byte, error) {
	var byteArr []byte
	substrs := strings.Split(QName, ".")

	for _, str := range substrs {
		l := uint64(len(str))

		byteArr = binary.AppendUvarint(byteArr, l)
		byteArr, _ = binary.Append(byteArr, binary.BigEndian, []byte(str))
	}
	byteArr = binary.AppendUvarint(byteArr, 0)
	return byteArr, nil
}

// IsSet checks if bit at pos is set or not. pos start from 0
func IsSet(b byte, pos uint) bool {
	var mask byte = (1 << (7 - pos))

	if (b & mask) != 0 {
		return true
	}
	return false

}

func DecodeName(msg []byte, start uint) (string, uint) {
	i := start
	var labels []string
	var jump bool
	var bytesUsed, bytesProcessed uint

	for bytesProcessed <= MaxDomainNameLength {
		// Check bits for pointer form
		if IsSet(msg[i], 0) && IsSet(msg[i], 1) {
			var offset uint16 = ((uint16(msg[i]) & 0x3F) << 8) | uint16(msg[i+1])
			start = uint(offset)
			i = start
			if jump == false {
				bytesUsed += 2
			}
			bytesProcessed += 2
			jump = true
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
		bytesProcessed += uint(len(label)) + 1
		labels = append(labels, label)

		i += (n + 1)
	}

	// Add last length octet if not jumped
	if jump == false {
		bytesUsed++
	}

	return strings.Join(labels, "."), bytesUsed
}
