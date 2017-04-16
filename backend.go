package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type responder func()

func respondWithFAIL() {
	fmt.Printf("FAIL\n")
}

func respondWithEND() {
	fmt.Printf("END\n")
}

//RunBackend This function implements the PowerDNS-Pipe-Backend protocol and generates
// the response data it possible
func RunBackend(conn *RedisConnection) {
	bio := bufio.NewReader(os.Stdin)

	// handshake with PowerDNS
	_, _, _ = bio.ReadLine()
	fmt.Printf("OK\tDDNS Go Backend\n")

	for {
		line, _, err := bio.ReadLine()
		if err != nil {
			respondWithFAIL()
			continue
		}

		HandleRequest(string(line), conn)()
	}
}

func HandleRequest(line string, conn *RedisConnection) responder {
	if Verbose {
		fmt.Printf("LOG\t'%s'\n", line)
	}

	parts := strings.Split(line, "\t")
	if len(parts) != 6 {
		return respondWithFAIL
	}

	queryName := parts[1]
	queryClass := parts[2]
	queryType := parts[3]
	queryID := parts[4]

	var response, record string
	record = queryType

	switch queryType {
	case "SOA":
		response = fmt.Sprintf("%s. hostmaster.example.com. %d 1800 3600 7200 5",
			DdnsSoaFqdn, getSoaSerial())

	case "NS":
		response = fmt.Sprintf("%s.", DdnsSoaFqdn)

	case "A":
	case "ANY":
		// get the host part of the fqdn: pi.d.example.org -> pi
		var hostname string
		if strings.HasSuffix(queryName, DdnsDomain) {
			hostname = queryName[:len(queryName)-len(DdnsDomain)]
		}

		if hostname == "" || !conn.HostExist(hostname) {
			return respondWithFAIL
		}

		host := conn.GetHost(hostname)
		response = host.IP

		record = "A"
		if !host.IsIPv4() {
			record = "AAAA"
		}

	default:
		return respondWithFAIL
	}

	fmt.Printf("DATA\t%s\t%s\t%s\t10\t%s\t%s\n",
		queryName, queryClass, record, queryID, response)
	return respondWithEND
}

func getSoaSerial() int64 {
	// return current time in milliseconds
	return time.Now().UnixNano()
}
