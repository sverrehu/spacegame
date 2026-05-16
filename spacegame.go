package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/sverrehu/goutils/getopt"
	"github.com/sverrehu/spacegame/internal/client"
	"github.com/sverrehu/spacegame/internal/server"
)

func main() {
	help := false
	serverPort := -1
	connectString := ""
	name := ""
	opts := []getopt.Option{
		{ShortName: 'h', LongName: "help", Type: getopt.Flag, Target: &help},
		{ShortName: 's', LongName: "server", Type: getopt.Integer, Target: &serverPort},
		{ShortName: 'c', LongName: "client", Type: getopt.String, Target: &connectString},
		{ShortName: 'n', LongName: "name", Type: getopt.String, Target: &name},
	}
	getopt.Parse(&os.Args, opts, false)
	if help || serverPort < 0 && connectString == "" {
		usage()
	}
	if serverPort >= 0 {
		server.StartServer(serverPort)
	}
	if connectString != "" {
		host, portString, err := net.SplitHostPort(connectString)
		if err != nil {
			panic(err)
		}
		port, err := strconv.Atoi(portString)
		if err != nil {
			panic(err)
		}
		if name == "" {
			fmt.Println("You must specify a name")
			os.Exit(1)
		}
		client := client.NewTcpClient(host, port, name)
		client.Start()
	}
	if serverPort >= 0 {
		server.WaitForServer()
	}
}

func usage() {
	fmt.Println("Usage: spacegame option ...")
	fmt.Println("")
	fmt.Println("  -s, --server=PORT        start server at given port")
	fmt.Println("  -c, --connect=HOST:PORT  start client")
	fmt.Println("  -n, --name=NAME          player name, required for client")
	fmt.Println("")
	os.Exit(0)
}
