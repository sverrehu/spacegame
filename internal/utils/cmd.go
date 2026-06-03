package utils

import (
	"net"
	"strconv"
)

func HostAndPort(connectString string) (string, int) {
	host, portString, err := net.SplitHostPort(connectString)
	if err != nil {
		panic(err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		panic(err)
	}
	return host, port
}
