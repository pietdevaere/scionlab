// sensorfetcher application
// For documentation on how to setup and run the application see:
// https://github.com/perrig/scionlab/blob/master/README.md
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"os"

	"github.com/scionproto/scion/go/lib/snet"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func printUsage() {
	fmt.Println("scion-sensor-server -s ServerSCIONAddress -c ClientSCIONAddress")
	fmt.Println("The SCION address is specified as ISD-AS,[IP Address]:Port")
	fmt.Println("Example SCION address 1-1,[127.0.0.1]:42002")
}

func main() {
	var (
		clientAddress string
		serverAddress string

		csvPath string

		err    error
		local  *snet.Addr
		remote *snet.Addr

		udpConnection *snet.Conn
	)

	// Fetch arguments from command line
	flag.StringVar(&csvPath, "f", "", "File to write data to")
	flag.StringVar(&clientAddress, "c", "", "Client SCION Address")
	flag.StringVar(&serverAddress, "s", "", "Server SCION Address")
	flag.Parse()

	// Create the SCION UDP socket
	if len(clientAddress) > 0 {
		local, err = snet.AddrFromString(clientAddress)
		check(err)
	} else {
		printUsage()
		check(fmt.Errorf("Error, client address needs to be specified with -c"))
	}
	if len(serverAddress) > 0 {
		remote, err = snet.AddrFromString(serverAddress)
		check(err)
	} else {
		printUsage()
		check(fmt.Errorf("Error, server address needs to be specified with -s"))
	}

	sciondAddr := "/run/shm/sciond/sd" + strconv.Itoa(local.IA.I) + "-" + strconv.Itoa(local.IA.A) + ".sock"
	dispatcherAddr := "/run/shm/dispatcher/default.sock"
	snet.Init(local.IA, sciondAddr, dispatcherAddr)

	udpConnection, err = snet.DialSCION("udp4", local, remote)
	check(err)

	receivePacketBuffer := make([]byte, 2500)
	sendPacketBuffer := make([]byte, 0)

	n, err := udpConnection.Write(sendPacketBuffer)
	check(err)

	n, _, err = udpConnection.ReadFrom(receivePacketBuffer)
	check(err)

	raw_response := string(receivePacketBuffer[:n])
	split_response := strings.Split(raw_response, "\n")

	fmt.Println(raw_response)

	response := make(map[string]string)

	response["date"] = split_response[0]

	for i:=1; i<len(split_response); i++ {
		split_line := strings.Split(split_response[i], ": ")
		if len(split_line) > 1 {
			response[split_line[0]] = split_line[1]
//			fmt.Println(split_line[0], "is", split_line[1])
		}
	}

	// fmt.Println(response)

	columns := []string{"date", "Illuminance", "UV Light", "CO2", "Sound intensity", "Humidity", "Temperature", "Motion"}

	if len(csvPath) > 0 {
		csvFile, err := os.OpenFile(csvPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		check(err)

		for i, column := range columns {
			fmt.Fprintf(csvFile, "%v", response[column])
			if i == len(columns) - 1{
				fmt.Fprintf(csvFile, "\n")
			} else {
				fmt.Fprintf(csvFile, ",")
			}
		}
	}
}

