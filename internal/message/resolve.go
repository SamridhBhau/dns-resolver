package message

import (
	"fmt"
)

const RootNameServer = "192.33.4.12:53"

func (m Message) Resolve(address string) (string, error) {
	fmt.Printf("Querying " + address + " for " + m.Q.QName + "\n")

	recvBytes, err := m.SendRequest(address)

	if err != nil {
		fmt.Println("Send request failed")
		return "", err
	}

	recvMsg := ParseResponse(recvBytes)

	if recvMsg.H.ANCOUNT > 0 {
		answer := recvMsg.ANS[0]
		return answer.RData, nil
	}
	if recvMsg.H.NSCOUNT > 0 {
		var found bool
		nameServerIP := ""
		for _, ns := range recvMsg.AUTH {
			if ns.Type == 2 {
				for _, ad := range recvMsg.ADD {
					if ad.Type == 1 && ad.Name == ns.RData {
						found = true
						nameServerIP = ad.RData
						break
					}
				}
			}
		}

		if found {
			res, err := m.Resolve(nameServerIP + ":53")

			if err != nil {
				return "", err
			}

			return res, nil

		} else {
			var nameServer ResourceRecord
			var found bool
			for _, ns := range recvMsg.AUTH {
				if ns.Type == 2 {
					nameServer = ns
					found = true
					break
				}
			}

			if found {
				newMsg := Message{
					H: Header{
						ID:      22,
						QDCOUNT: 1,
					},
					Q: Question{
						nameServer.RData,
						1,
						1,
					},
				}

				nameServerIP, err := newMsg.Resolve(RootNameServer)

				if err != nil {
					fmt.Println("Failed to resolve: " + nameServer.RData)
					return "", nil
				}

				res, err := m.Resolve(nameServerIP + ":53")

				if err != nil {
					fmt.Println("Failed to resolve: " + m.Q.QName)
					return "", nil
				}

				return res, nil
			}
		}
	}
	return "", fmt.Errorf("not found")
}
